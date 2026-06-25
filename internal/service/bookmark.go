package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/bookmarkurl"
	"pet-link/internal/pkg/cardquality"
	"pet-link/internal/pkg/gemini"
	"pet-link/internal/pkg/pagemeta"
)

type BookmarkRepository interface {
	Create(ctx context.Context, userID string, input domain.CreateBookmarkInput) (domain.Bookmark, error)
	ExistsByURLForUser(ctx context.Context, userID, canonicalURL string) (bool, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Bookmark, error)
	GetByIDForUser(ctx context.Context, userID, bookmarkID string) (domain.Bookmark, error)
	UpdateImageURL(ctx context.Context, userID, bookmarkID, imageURL string) error
	UpdateEnrichment(ctx context.Context, userID, bookmarkID string, enrichment domain.BookmarkEnrichment) error
	MarkEnriched(ctx context.Context, userID, bookmarkID string) error
	Delete(ctx context.Context, userID, bookmarkID string) error
}

type BookmarkEnricher interface {
	Enrich(ctx context.Context, rawURL string) (domain.BookmarkEnrichment, error)
	Classify(ctx context.Context, pageURL, title, description string) (domain.BookmarkEnrichment, error)
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

const (
	imageFetchTimeout    = 16 * time.Second
	quickMetaTimeout     = 2 * time.Second
	metaFallbackTimeout  = 12 * time.Second
	enrichAttemptTimeout = 20 * time.Second
	enrichTotalTimeout   = 4 * time.Minute
	enrichMaxAttempts    = 8
	enrichRetryBaseDelay = 2 * time.Second
	enrichRetryMaxDelay  = 30 * time.Second
)

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

	s.applyQuickFallback(ctx, &input)

	if input.ImageURL == "" {
		if thumb := pagemeta.PlatformThumbnailURL(input.URL); thumb != "" {
			input.ImageURL = thumb
		}
	}

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

	s.enrichAsync(bookmark.UserID, bookmark.ID, bookmark.URL)

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

func (s *bookmarkService) applyQuickFallback(ctx context.Context, input *domain.CreateBookmarkInput) {
	if s.metaFallback == nil {
		return
	}

	quickCtx, cancel := context.WithTimeout(ctx, quickMetaTimeout)
	defer cancel()

	fallback, ok := s.metaFallback.FallbackEnrich(quickCtx, input.URL)
	if !ok {
		return
	}

	s.applyEnrichment(input, gemini.NormalizeEnrichment(fallback))
}

func (s *bookmarkService) enrichAsync(userID, bookmarkID, rawURL string) {
	if s.enricher == nil && s.metaFallback == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), enrichTotalTimeout)
		defer cancel()

		hints, ok := s.loadEnrichmentHints(ctx, userID, bookmarkID)
		if !ok {
			return
		}

		defer func() {
			markCtx, markCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer markCancel()
			if err := s.repo.MarkEnriched(markCtx, userID, bookmarkID); err != nil && !errors.Is(err, domain.ErrBookmarkNotFound) {
				log.Printf("bookmark mark enriched failed for %s: %v", rawURL, err)
			}
		}()

		for attempt := 1; attempt <= enrichMaxAttempts; attempt++ {
			if s.enricher != nil {
				attemptCtx, attemptCancel := context.WithTimeout(ctx, enrichAttemptTimeout)
				enrichment, err := s.enricher.Enrich(attemptCtx, rawURL)
				attemptCancel()

				if err == nil && !gemini.IsUnavailableEnrichment(enrichment) {
					finished := s.finishEnrichment(ctx, rawURL, hints, enrichment)
					imageURL := s.bookmarkImageURL(ctx, userID, bookmarkID)
					s.tryPersistEnrichment(ctx, userID, bookmarkID, finished)
					if cardquality.IsGoodEnough(finished, imageURL) {
						return
					}
					hints = finished
				}

				if err != nil {
					log.Printf("bookmark enrich attempt %d failed for %s: %v", attempt, rawURL, err)
				}
			}

			if attempt == enrichMaxAttempts {
				break
			}

			delay := enrichRetryDelay(attempt)
			select {
			case <-ctx.Done():
				log.Printf("bookmark enrich timed out for %s", rawURL)
				return
			case <-time.After(delay):
			}
		}

		s.finalizeEnrichment(ctx, userID, bookmarkID, rawURL, hints)
	}()
}

func (s *bookmarkService) loadEnrichmentHints(ctx context.Context, userID, bookmarkID string) (domain.BookmarkEnrichment, bool) {
	bookmark, err := s.repo.GetByIDForUser(ctx, userID, bookmarkID)
	if err != nil {
		return domain.BookmarkEnrichment{}, false
	}
	return enrichmentFromBookmark(bookmark), true
}

func (s *bookmarkService) finishEnrichment(ctx context.Context, rawURL string, hints, enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	merged := cardquality.Merge(hints, gemini.NormalizeEnrichment(enrichment))
	if !needsClassification(merged) {
		return merged
	}
	return s.classifyAndMerge(ctx, rawURL, merged)
}

func (s *bookmarkService) finalizeEnrichment(ctx context.Context, userID, bookmarkID, rawURL string, hints domain.BookmarkEnrichment) {
	merged := hints

	metaCtx, metaCancel := context.WithTimeout(ctx, metaFallbackTimeout)
	defer metaCancel()

	if s.metaFallback != nil {
		if fallback, ok := s.metaFallback.FallbackEnrich(metaCtx, rawURL); ok {
			merged = cardquality.Merge(merged, gemini.NormalizeEnrichment(fallback))
		}
	}

	if merged.Title != "" || merged.Description != "" {
		merged = s.classifyAndMerge(ctx, rawURL, merged)
	}

	imageURL := s.bookmarkImageURL(ctx, userID, bookmarkID)
	s.tryPersistEnrichment(ctx, userID, bookmarkID, merged)

	if !cardquality.IsAcceptable(merged, imageURL) {
		log.Printf("bookmark enrich exhausted for %s", rawURL)
	}
}

func (s *bookmarkService) bookmarkImageURL(ctx context.Context, userID, bookmarkID string) string {
	bookmark, err := s.repo.GetByIDForUser(ctx, userID, bookmarkID)
	if err != nil {
		return ""
	}
	return bookmark.ImageURL
}

func (s *bookmarkService) classifyAndMerge(ctx context.Context, rawURL string, base domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	if s.enricher == nil || !needsClassification(base) {
		return base
	}

	classifyCtx, cancel := context.WithTimeout(ctx, enrichAttemptTimeout)
	defer cancel()

	classified, err := s.enricher.Classify(classifyCtx, rawURL, base.Title, base.Description)
	if err != nil {
		log.Printf("bookmark classify failed for %s: %v", rawURL, err)
		return base
	}

	if gemini.IsUnavailableEnrichment(classified) {
		return base
	}

	return cardquality.Merge(base, gemini.NormalizeEnrichment(classified))
}

func enrichmentFromBookmark(bookmark domain.Bookmark) domain.BookmarkEnrichment {
	return domain.BookmarkEnrichment{
		Title:       bookmark.Title,
		Description: bookmark.Description,
		Category:    bookmark.Category,
		Tags:        bookmark.Tags,
	}
}

func mergeEnrichment(base, patch domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	return cardquality.Merge(base, patch)
}

func needsClassification(enrichment domain.BookmarkEnrichment) bool {
	if enrichment.Title == "" && enrichment.Description == "" {
		return false
	}
	if len(enrichment.Tags) >= 2 && enrichment.Category != "" && enrichment.Category != "other" {
		return false
	}
	return true
}

func (s *bookmarkService) tryPersistEnrichment(ctx context.Context, userID, bookmarkID string, enrichment domain.BookmarkEnrichment) bool {
	if err := s.persistEnrichment(ctx, userID, bookmarkID, enrichment); err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return true
		}
		log.Printf("bookmark enrich update failed: %v", err)
		return false
	}
	return true
}

func enrichRetryDelay(attempt int) time.Duration {
	delay := enrichRetryBaseDelay * time.Duration(1<<(attempt-1))
	if delay > enrichRetryMaxDelay {
		return enrichRetryMaxDelay
	}
	return delay
}

func (s *bookmarkService) persistEnrichment(ctx context.Context, userID, bookmarkID string, enrichment domain.BookmarkEnrichment) error {
	enrichment = gemini.NormalizeEnrichment(enrichment)
	if enrichment.Title == "" && enrichment.Description == "" && enrichment.Category == "" && len(enrichment.Tags) == 0 {
		return nil
	}
	return s.repo.UpdateEnrichment(ctx, userID, bookmarkID, enrichment)
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
