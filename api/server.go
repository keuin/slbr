package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keuin/slbr/api/agent"
)

func StartServer(addr string, a agent.Agent) error {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/tasks", func(c *fiber.Ctx) error {
		return c.JSON(a.GetTasks())
	})

	return app.Listen(addr)
}
