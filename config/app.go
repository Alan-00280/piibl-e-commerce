package config

import (
	"errors"
	"log/slog"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/Alan-00280/piibl-e-commerce.git/middleware"
	"github.com/Alan-00280/piibl-e-commerce.git/routes"
	"github.com/gofiber/fiber/v2"
)

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)

		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_field",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", appErr.Unwrap().Error()),
			)
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
			)
		}

		return c.Status(appErr.Status).JSON(model.ErrorRespone{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}

func NewApp(logger *slog.Logger, deps routes.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "api-backend"),
		ErrorHandler: newErrorHandler(logger),
	})

	// Middleware
	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", "http://localhost:5173/"))

	// HELLO WORLD!
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// App routess
	routes.Register(app, deps)

	// 404
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("404 - routes not found")
	})

	return app
}
