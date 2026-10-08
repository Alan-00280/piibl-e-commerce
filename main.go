package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/repository"
	"github.com/Alan-00280/piibl-e-commerce.git/app/service"
	"github.com/Alan-00280/piibl-e-commerce.git/config"
	"github.com/Alan-00280/piibl-e-commerce.git/database"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/Alan-00280/piibl-e-commerce.git/routes"
)

var minJwtSecretLength = 32

func main() {
	// LOAD ENV
	config.LoadEnv()

	// LOGGER CONFIG
	logger := config.NewLogger()

	// JWT SECRET
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minJwtSecretLength {
		logger.Error("jwt secret isn't valid please check again", slog.Int("min length:", minJwtSecretLength))
		os.Exit(1)
	}

	// DB POOL
	pool, err := database.NewPool(context.Background())
	if err != nil {
		fmt.Printf("CAN'T CREATE DB POOL: %v", err)
		os.Exit(1)
		return
	}
	defer pool.Close()

	// JWT MANAGER
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "be-prak"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	// LOAD ROLE PERMISSIONS
	roleRepo := repository.NewRoleRepository(pool)
	rawRolePerm, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat role dan permissions", slog.String("error", err.Error()))
		os.Exit(1)
	}

	permissionSet := helper.NewPermissionSet(rawRolePerm)
	logger.Info("berhasil memuat roles", slog.Any("roles", permissionSet.KnownRoles()))

	// COMMON PASSWORD
	passwordCommonPath, err := filepath.Abs("./files/common_password.txt")
	if err != nil {
		logger.Error("gagal memuat password umum", slog.String("error", err.Error()))
		os.Exit(1)
	}

	passwordCommonSet, err := helper.NewPasswordCommonSet(passwordCommonPath)
	if err != nil {
		logger.Error("gagal memuat password umum", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// APP VALIDATOR
	appValidator := helper.NewValidator(passwordCommonSet)

	// REPO & SERVICES
	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo, permissionSet, appValidator)

	authRepo := repository.NewAuthRepo(pool)
	authService := service.NewAuthService(
		userRepo,
		authRepo,
		jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
		appValidator,
		permissionSet,
	)

	storeRepo := repository.NewStoreRepository(pool)
	storeService := service.NewStoreService(storeRepo, permissionSet, appValidator)

	productRepo := repository.NewProductRepository(pool)
	productService := service.NewProductService(productRepo, storeRepo, permissionSet, appValidator)

	// APP
	deps := routes.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permission:     permissionSet,
		AuthService:    authService,
		UserService:    userService,
		StoreService:   storeService,
		ProductService: productService,
	}

	app := config.NewApp(logger, deps)
	port := config.GetEnv("APP_PORT", "3000")

	// RUNNING THE APP
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("SERVER STOPPED", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("SERVER RUNNING", slog.String("PORT", port))

	// GRACEFULL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("SHUTTING DOWN SERVER...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("FAIL TO SHUTDOWN SERVER", slog.String("ERROR", err.Error()))
	}

	logger.Info("SERVER SHUTTED DOWN...")

}
