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

func OkList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func FailValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false,
		Message: "validation fail",
		Errors:  errors,
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func SuccessCursor(c *fiber.Ctx, message string, rows any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponseCursor{
		Success: true,
		Message: message,
		Data:    rows,
		Meta:    meta,
	})
}
