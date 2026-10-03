package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/config"
	"github.com/Alan-00280/piibl-e-commerce.git/database"
	"github.com/Alan-00280/piibl-e-commerce.git/routes"
)

func main() {
	// LOAD ENV
	config.LoadEnv()

	// LOGGER CONFIG
	logger := config.NewLogger()

	// DB POOL
	pool, err := database.NewPool(context.Background())
	if err != nil {
		fmt.Printf("CAN'T CREATE DB POOL: %v", err)
		os.Exit(1)
		return
	}
	defer pool.Close()

	// APP
	deps := routes.Dependencies{
		Pool: pool,
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
