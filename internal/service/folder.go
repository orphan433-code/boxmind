package service

import (
	"context"
	"fmt"
	"strings"

	"pet-link/internal/domain"
)

type FolderRepository interface {
	Create(ctx context.Context, userID string, input domain.CreateFolderInput) (domain.Folder, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Folder, error)
	GetByIDForUser(ctx context.Context, userID, folderID string) (domain.Folder, error)
	UpdateName(ctx context.Context, userID, folderID, name string) (domain.Folder, error)
	Delete(ctx context.Context, userID, folderID string) error
}

type FolderService interface {
	Create(ctx context.Context, userID string, input domain.CreateFolderInput) (domain.Folder, error)
	List(ctx context.Context, userID string) ([]domain.Folder, error)
	Update(ctx context.Context, userID, folderID string, input domain.UpdateFolderInput) (domain.Folder, error)
	Delete(ctx context.Context, userID, folderID string) error
	AssignBookmark(ctx context.Context, userID, bookmarkID, folderID string) (domain.Bookmark, error)
}

type BookmarkFolderUpdater interface {
	UpdateFolderID(ctx context.Context, userID, bookmarkID, folderID string) (domain.Bookmark, error)
}

type folderService struct {
	repo         FolderRepository
	bookmarkRepo BookmarkFolderUpdater
}

func NewFolderService(repo FolderRepository) FolderService {
	return &folderService{repo: repo}
}

func NewFolderServiceWithBookmarks(repo FolderRepository, bookmarkRepo BookmarkFolderUpdater) FolderService {
	return &folderService{repo: repo, bookmarkRepo: bookmarkRepo}
}

func (s *folderService) Create(ctx context.Context, userID string, input domain.CreateFolderInput) (domain.Folder, error) {
	name, err := normalizeFolderName(input.Name)
	if err != nil {
		return domain.Folder{}, err
	}
	if userID == "" {
		return domain.Folder{}, fmt.Errorf("user id is required")
	}

	return s.repo.Create(ctx, userID, domain.CreateFolderInput{Name: name})
}

func (s *folderService) List(ctx context.Context, userID string) ([]domain.Folder, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListByUserID(ctx, userID)
}

func (s *folderService) Update(ctx context.Context, userID, folderID string, input domain.UpdateFolderInput) (domain.Folder, error) {
	name, err := normalizeFolderName(input.Name)
	if err != nil {
		return domain.Folder{}, err
	}
	if userID == "" {
		return domain.Folder{}, fmt.Errorf("user id is required")
	}
	if folderID == "" {
		return domain.Folder{}, fmt.Errorf("folder id is required")
	}

	return s.repo.UpdateName(ctx, userID, folderID, name)
}

func (s *folderService) Delete(ctx context.Context, userID, folderID string) error {
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	if folderID == "" {
		return fmt.Errorf("folder id is required")
	}
	return s.repo.Delete(ctx, userID, folderID)
}

func (s *folderService) AssignBookmark(ctx context.Context, userID, bookmarkID, folderID string) (domain.Bookmark, error) {
	if userID == "" {
		return domain.Bookmark{}, fmt.Errorf("user id is required")
	}
	if bookmarkID == "" {
		return domain.Bookmark{}, fmt.Errorf("bookmark id is required")
	}
	if s.bookmarkRepo == nil {
		return domain.Bookmark{}, fmt.Errorf("bookmark assignment is unavailable")
	}

	folderID = strings.TrimSpace(folderID)
	if folderID != "" {
		if _, err := s.repo.GetByIDForUser(ctx, userID, folderID); err != nil {
			return domain.Bookmark{}, err
		}
	}

	return s.bookmarkRepo.UpdateFolderID(ctx, userID, bookmarkID, folderID)
}

func normalizeFolderName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", fmt.Errorf("folder name is required")
	}
	if len([]rune(name)) > 80 {
		return "", fmt.Errorf("folder name is too long")
	}
	return name, nil
}
