package routes

import (
	"context"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool *pgxpool.Pool
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// HEALTH CHECK (WITH DB POOL CONNECTION)
	api.Get("/health", healthCheck(deps.Pool))

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
