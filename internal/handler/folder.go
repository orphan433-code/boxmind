package handler

import (
	"errors"

	"pet-link/internal/domain"
	"pet-link/internal/middleware"
	"pet-link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type FolderHandler struct {
	folderService service.FolderService
}

func NewFolderHandler(folderService service.FolderService) *FolderHandler {
	return &FolderHandler{folderService: folderService}
}

type createFolderRequest struct {
	Name string `json:"name"`
}

type updateFolderRequest struct {
	Name string `json:"name"`
}

func (h *FolderHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req createFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	folder, err := h.folderService.Create(c.Context(), userID, domain.CreateFolderInput{Name: req.Name})
	if err != nil {
		if errors.Is(err, domain.ErrFolderAlreadyExists) {
			return JSON(c, fiber.StatusConflict, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusCreated, folder)
}

func (h *FolderHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	folders, err := h.folderService.List(c.Context(), userID)
	if err != nil {
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	if folders == nil {
		folders = []domain.Folder{}
	}

	return JSON(c, fiber.StatusOK, folders)
}

func (h *FolderHandler) Update(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req updateFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	folder, err := h.folderService.Update(c.Context(), userID, c.Params("id"), domain.UpdateFolderInput{Name: req.Name})
	if err != nil {
		if errors.Is(err, domain.ErrFolderNotFound) {
			return JSON(c, fiber.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, domain.ErrFolderAlreadyExists) {
			return JSON(c, fiber.StatusConflict, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusOK, folder)
}

func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	if err := h.folderService.Delete(c.Context(), userID, c.Params("id")); err != nil {
		if errors.Is(err, domain.ErrFolderNotFound) {
			return JSON(c, fiber.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type assignBookmarkFolderRequest struct {
	FolderID *string `json:"folder_id"`
}

func (h *FolderHandler) AssignBookmark(c *fiber.Ctx) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return JSON(c, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req assignBookmarkFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: "invalid json"})
	}

	folderID := ""
	if req.FolderID != nil {
		folderID = *req.FolderID
	}

	bookmark, err := h.folderService.AssignBookmark(c.Context(), userID, c.Params("bookmarkId"), folderID)
	if err != nil {
		if errors.Is(err, domain.ErrFolderNotFound) || errors.Is(err, domain.ErrBookmarkNotFound) {
			return JSON(c, fiber.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		return JSON(c, fiber.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return JSON(c, fiber.StatusOK, bookmark)
}
