package helper

import (
	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/gofiber/fiber/v2"
)

func Ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}
