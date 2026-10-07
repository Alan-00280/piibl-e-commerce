package middleware

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		requestID, _ := c.Locals("requestid").(string)

		status := c.Response().StatusCode()
		if err != nil {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		attr := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		// if user, ok := helper.CurentUser(c); ok {
		// 	attr = append(attr,
		// 		slog.Int("user_id", user.UserID),
		// 		slog.String("role", user.Role),
		// 	)
		// }

		logger.Info("http_request", attr...)

		return err
	}
}

var bodiedMethod = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if bodiedMethod[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.UnsuportedMediaType("Content-Type harus application/json")
		}
	}
	return c.Next()
}

func Register(app *fiber.App, logger *slog.Logger, allowedOrigin string) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(corsPolicy(allowedOrigin))
	app.Use(RequestLogger(logger))
}

// func corsPolicy(allowedOrigin string) fiber.Handler : membatasi origin pemanggil API
func corsPolicy(allowedOrigin string) fiber.Handler {
	if strings.TrimSpace(allowedOrigin) == "" {
		allowedOrigin = "http://localhost:5173/"
	}

	return cors.New(cors.Config{
		AllowOrigins: allowedOrigin,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	})
}
