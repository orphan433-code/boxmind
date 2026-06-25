package handler

import "github.com/gofiber/fiber/v2"

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}
