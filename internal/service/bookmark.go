package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/bookmarkurl"
	"pet-link/internal/pkg/gemini"
)

type BookmarkRepository interface {
	Create(ctx context.Context, userID string, input domain.CreateBookmarkInput) (domain.Bookmark, error)
	ExistsByURLForUser(ctx context.Context, userID, canonicalURL string) (bool, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Bookmark, error)
	GetByIDForUser(ctx context.Context, userID, bookmarkID string) (domain.Bookmark, error)
	UpdateImageURL(ctx context.Context, userID, bookmarkID, imageURL string) error
	Delete(ctx context.Context, userID, bookmarkID string) error
}

type BookmarkEnricher interface {
	Enrich(ctx context.Context, rawURL string) (domain.BookmarkEnrichment, error)
}

type BookmarkImageFetcher interface {
	FetchImageURL(ctx context.Context, rawURL string) (string, error)
}

type BookmarkService interface {
	Create(ctx context.Context, userID string, input domain.CreateBookmarkInput) (domain.Bookmark, error)
	List(ctx context.Context, userID string) ([]domain.Bookmark, error)
	GetByID(ctx context.Context, userID, bookmarkID string) (domain.Bookmark, error)
	Delete(ctx context.Context, userID, bookmarkID string) error
}

const imageFetchTimeout = 4 * time.Second

type BookmarkMetaFallback interface {
	FallbackEnrich(ctx context.Context, rawURL string) (domain.BookmarkEnrichment, bool)
}

type bookmarkService struct {
	repo         BookmarkRepository
	enricher     BookmarkEnricher
	imageFetcher BookmarkImageFetcher
	metaFallback BookmarkMetaFallback
}

func NewBookmarkService(
	repo BookmarkRepository,
	enricher BookmarkEnricher,
	imageFetcher BookmarkImageFetcher,
	metaFallback BookmarkMetaFallback,
) BookmarkService {
	return &bookmarkService{
		repo:         repo,
		enricher:     enricher,
		imageFetcher: imageFetcher,
		metaFallback: metaFallback,
	}
}

func (s *bookmarkService) Create(ctx context.Context, userID string, input domain.CreateBookmarkInput) (domain.Bookmark, error) {
	input.URL = strings.TrimSpace(input.URL)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.Category = strings.TrimSpace(input.Category)

	if userID == "" {
		return domain.Bookmark{}, fmt.Errorf("user id is required")
	}
	if input.URL == "" {
		return domain.Bookmark{}, fmt.Errorf("url is required")
	}

	canonicalURL, err := bookmarkurl.Normalize(input.URL)
	if err != nil {
		return domain.Bookmark{}, fmt.Errorf("invalid url")
	}
	input.URL = canonicalURL

	exists, err := s.repo.ExistsByURLForUser(ctx, userID, canonicalURL)
	if err != nil {
		return domain.Bookmark{}, err
	}
	if exists {
		return domain.Bookmark{}, domain.ErrBookmarkAlreadyExists
	}

	s.enrichFromAI(ctx, &input)

	if input.Title == "" {
		input.Title = input.URL
	}
	if input.Category == "" {
		input.Category = "other"
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	bookmark, err := s.repo.Create(ctx, userID, input)
	if err != nil {
		return domain.Bookmark{}, err
	}

	if bookmark.ImageURL == "" {
		s.fetchImageAsync(bookmark.UserID, bookmark.ID, bookmark.URL)
	}

	return bookmark, nil
}

func (s *bookmarkService) fetchImageAsync(userID, bookmarkID, rawURL string) {
	if s.imageFetcher == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), imageFetchTimeout)
		defer cancel()

		imageURL, err := s.imageFetcher.FetchImageURL(ctx, rawURL)
		if err != nil {
			log.Printf("bookmark image fetch failed for %s: %v", rawURL, err)
			return
		}
		if imageURL == "" {
			return
		}

		if err := s.repo.UpdateImageURL(ctx, userID, bookmarkID, imageURL); err != nil {
			log.Printf("bookmark image update failed for %s: %v", rawURL, err)
		}
	}()
}

func (s *bookmarkService) enrichFromAI(ctx context.Context, input *domain.CreateBookmarkInput) {
	if s.enricher == nil {
		return
	}

	enrichment, err := s.enricher.Enrich(ctx, input.URL)
	if err != nil {
		log.Printf("bookmark enrich failed for %s: %v", input.URL, err)
		s.applyMetaFallback(ctx, input)
		return
	}

	if gemini.IsUnavailableEnrichment(enrichment) {
		if s.applyMetaFallback(ctx, input) {
			return
		}
	}

	s.applyEnrichment(input, enrichment)
}

func (s *bookmarkService) applyMetaFallback(ctx context.Context, input *domain.CreateBookmarkInput) bool {
	if s.metaFallback == nil {
		return false
	}

	fallback, ok := s.metaFallback.FallbackEnrich(ctx, input.URL)
	if !ok {
		return false
	}

	s.applyEnrichment(input, gemini.NormalizeEnrichment(fallback))
	return true
}

func (s *bookmarkService) applyEnrichment(input *domain.CreateBookmarkInput, enrichment domain.BookmarkEnrichment) {
	if input.Title == "" {
		input.Title = enrichment.Title
	}
	if input.Description == "" {
		input.Description = enrichment.Description
	}
	if input.Category == "" {
		input.Category = enrichment.Category
	}
	if len(input.Tags) == 0 && len(enrichment.Tags) > 0 {
		input.Tags = enrichment.Tags
	}
}

func (s *bookmarkService) List(ctx context.Context, userID string) ([]domain.Bookmark, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListByUserID(ctx, userID)
}

func (s *bookmarkService) GetByID(ctx context.Context, userID, bookmarkID string) (domain.Bookmark, error) {
	if userID == "" || bookmarkID == "" {
		return domain.Bookmark{}, fmt.Errorf("user id and bookmark id are required")
	}
	return s.repo.GetByIDForUser(ctx, userID, bookmarkID)
}

func (s *bookmarkService) Delete(ctx context.Context, userID, bookmarkID string) error {
	if userID == "" || bookmarkID == "" {
		return fmt.Errorf("user id and bookmark id are required")
	}
	return s.repo.Delete(ctx, userID, bookmarkID)
}
