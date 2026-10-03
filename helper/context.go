package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

const LocalsAuthUser = "authUser"

// func CurentUser(c *fiber.Ctx) (model.AuthUser, bool) {
// 	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
// 	return user, ok
// }

func RequestID(c *fiber.Ctx) string {
	reqID, _ := c.Locals("requestid").(string)
	return reqID
}

func ReqContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}
