package handler

import (
	"errors"

	"pet-link/internal/domain"
	"pet-link/internal/middleware"
	"pet-link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Me(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		}
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	return JSON(c, fiber.StatusOK, user)
}
