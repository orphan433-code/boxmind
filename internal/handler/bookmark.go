package handler

import (
	"errors"

	"pet-link/internal/domain"
	"pet-link/internal/middleware"
	"pet-link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BookmarkHandler struct {
	bookmarkService service.BookmarkService
}

func NewBookmarkHandler(bookmarkService service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{bookmarkService: bookmarkService}
}

type createBookmarkRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *BookmarkHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req createBookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	bookmark, err := h.bookmarkService.Create(c.Context(), userID, domain.CreateBookmarkInput{
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkAlreadyExists) {
			return JSON(c, fiber.StatusConflict, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusCreated, bookmark)
}

func (h *BookmarkHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	bookmarks, err := h.bookmarkService.List(c.Context(), userID)
	if err != nil {
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	return JSON(c, fiber.StatusOK, bookmarks)
}

func (h *BookmarkHandler) GetByID(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	bookmark, err := h.bookmarkService.GetByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return JSON(c, fiber.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	return JSON(c, fiber.StatusOK, bookmark)
}

func (h *BookmarkHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	if err := h.bookmarkService.Delete(c.Context(), userID, c.Params("id")); err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return JSON(c, fiber.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
