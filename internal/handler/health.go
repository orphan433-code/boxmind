package handler

import (
	"pet-link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	healthService service.HealthService
}

func NewHealthHandler(healthService service.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	status := h.healthService.Check(c.Context())
	httpStatus := fiber.StatusOK
	if status.Status == "degraded" {
		httpStatus = fiber.StatusServiceUnavailable // 503
	}
	return JSON(c, httpStatus, status)
}
