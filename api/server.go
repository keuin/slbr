package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keuin/slbr/logging"
)

func StartServer(logger logging.Logger, addr string) error {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	return app.Listen(addr)
}
