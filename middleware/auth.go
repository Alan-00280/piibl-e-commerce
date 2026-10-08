package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/Alan-00280/piibl-e-commerce.git/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// func RequireAuth(*helper.JWTManager) fiber.handler
// memeriksa ada access token atau enggak. Ada --> data user disimpan di Locals
func RequireAuth(jwt *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm=api`)
			return helper.Unauthorized("header Authorization gagal diproses")
		}

		authUser, err := jwt.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm=api`)

			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("akses token sudah kedaluarsa")
			}

			return helper.Unauthorized("token invalid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

// func bearerToken(*fiber.Ctx) (string, error): mengambil token dari header
func bearerToken(c *fiber.Ctx) (string, error) {
	auth_header := c.Get(fiber.HeaderAuthorization)
	if auth_header == "" {
		return "", errors.New("header authorization kosong")
	}

	parts := strings.SplitN(auth_header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format header bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}

	return token, nil
}

// func LoginRateLimiter() fiber.Handler
// untuk membatasi jumlah request untuk satu source ip yang sama pada route login
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			// c.Set("Retry-After (s)", "60")
			return helper.TooManyRequest("terlalu banyak percobaan login")
		},
	})
}

// TODO: Limiter Saat Checkout

