package routes

import (
	"context"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/app/service"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/Alan-00280/piibl-e-commerce.git/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	AuthService    *service.AuthService
	UserService    *service.UserService
	StoreService   *service.StoreService
	ProductService *service.ProductService
	OrderService   *service.OrderService
	JWT            *helper.JWTManager
	Permission     *helper.PermissionSet
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")
	my := api.Group("/my", middleware.RequireAuth(deps.JWT))
	permissions := deps.Permission

	// HEALTH CHECK (WITH DB POOL CONNECTION)
	api.Get("/health", healthCheck(deps.Pool))

	// AUTHENTICATION
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// USER
	user := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	user.Get("/", middleware.RequirePermission(permissions, "user:list"), deps.UserService.ListAll)
	user.Post("/", middleware.RequirePermission(permissions, "user:update:any"), deps.UserService.Create)
	user.Delete("/:id", middleware.RequirePermission(permissions, "user:delete"), deps.UserService.Delete)
	user.Patch("/:id/role", middleware.RequirePermission(permissions, "role:assign"), deps.UserService.AssignRole)

	user.Put("/:id", deps.UserService.Replace)
	user.Get("/:id", deps.UserService.Get)
	user.Patch("/:id", deps.UserService.Patch)

	// STORES
	store := api.Group("/stores", middleware.RequireJSON)
	store.Get("/", deps.StoreService.ListAll)
	store.Get("/:id", deps.StoreService.Get)

	store.Post("/", middleware.RequireAuth(deps.JWT), middleware.RequirePermission(permissions, "store:create"), deps.StoreService.Create)
	store.Patch("/:id", middleware.RequireAuth(deps.JWT), deps.StoreService.Patch)
	store.Delete("/:id", middleware.RequireAuth(deps.JWT), deps.StoreService.Deactivate)
	my.Get("/stores", deps.StoreService.GetStoreTenant)

	// PRODUCT
	product := api.Group("/products")
	product.Get("/", deps.ProductService.ListAll)
	product.Get("/:id", deps.ProductService.Get)
	product.Post("/", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(permissions, "product:create"), deps.ProductService.Create)
	product.Patch("/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.Patch)
	product.Delete("/:id", middleware.RequireAuth(deps.JWT), deps.ProductService.Deactivate)
	product.Patch("/:id/stock", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.Restock)
	product.Patch("/:id/price", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.Reprice)
	my.Get("/products", deps.ProductService.ProductStore)

	// VARIANT SPACES
	product.Post("/:id/variant-spaces", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.CreateSVariant)
	api.Patch("/variant-spaces/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.PatchSVariant)
	api.Delete("/variant-spaces/:id", middleware.RequireAuth(deps.JWT), deps.ProductService.DeleteSVaraint)

	// VARIANT
	api.Post("/variants", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.CreateVariant)
	api.Patch("/variants/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.ProductService.PatchVariant)
	api.Delete("/variants/:id", middleware.RequireAuth(deps.JWT), deps.ProductService.DeleteVariant)

	// CHECKOUT
	order := api.Group("/orders", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	order.Post("/checkout", middleware.GeneralRateLimiter(), middleware.RequirePermission(permissions, "order:create"), deps.OrderService.Checkout)

	// ORDERS
	order.Get("/", deps.OrderService.ListAll)
	order.Get("/:id", deps.OrderService.Get)
	order.Patch("/:id/status", deps.OrderService.UpdateStatus)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			// ERR 503 - Service Unavailable
			return helper.ServiceUnavailable("database can't be reached")
		}

		return helper.Ok(c, "server and database is OK!", nil)
	}
}
