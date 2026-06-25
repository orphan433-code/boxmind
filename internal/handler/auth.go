package handler

import (
	"errors"

	"pet-link/internal/domain"
	"pet-link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type loginRequest struct {
	Email string `json:"email"`
}

type verifyRequest struct {
	Email string     `json:"email"`
	Code  jsonString `json:"code"`
}

type loginResponse struct {
	Message string `json:"message"`
}

func (h *AuthHandler) RequestLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	if err := h.authService.RequestLogin(c.Context(), req.Email); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusOK, loginResponse{Message: "code sent"})
}

func (h *AuthHandler) VerifyLogin(c *fiber.Ctx) error {
	var req verifyRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	result, err := h.authService.VerifyLogin(c.Context(), req.Email, req.Code.String())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOTP) {
			return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusOK, result)
}
