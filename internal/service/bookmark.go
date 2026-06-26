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
	Classify(ctx context.Context, pageURL, title, description, titleSource string) (domain.BookmarkEnrichment, error)
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
	imageFetchTimeout     = 16 * time.Second
	imageFetchAttempts    = 3
	imageFetchRetryDelay  = 2 * time.Second
	quickMetaTimeout      = 2 * time.Second
	metaFallbackTimeout   = 12 * time.Second
	enrichAttemptTimeout  = 20 * time.Second
	enrichLoopTimeout     = 110 * time.Second
	enrichFinalizeTimeout = 45 * time.Second
	enrichMaxAttempts     = 4
	enrichRetryBaseDelay  = 2 * time.Second
	enrichRetryMaxDelay   = 15 * time.Second
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

	s.applyQuickContentFallback(ctx, &input)

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
		var lastErr error
		for attempt := 1; attempt <= imageFetchAttempts; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), imageFetchTimeout)
			imageURL, err := s.imageFetcher.FetchImageURL(ctx, rawURL)
			cancel()

			if err == nil && imageURL != "" {
				updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
				err := s.repo.UpdateImageURL(updateCtx, userID, bookmarkID, imageURL)
				updateCancel()
				if err != nil {
					log.Printf("bookmark image update failed for %s: %v", rawURL, err)
				}
				return
			}

			lastErr = err
			if attempt < imageFetchAttempts {
				time.Sleep(imageFetchRetryDelay * time.Duration(attempt))
			}
		}

		if lastErr != nil {
			log.Printf("bookmark image fetch failed for %s: %v", rawURL, lastErr)
		}
	}()
}

func (s *bookmarkService) applyQuickContentFallback(ctx context.Context, input *domain.CreateBookmarkInput) {
	if s.metaFallback == nil {
		return
	}

	quickCtx, cancel := context.WithTimeout(ctx, quickMetaTimeout)
	defer cancel()

	fallback, ok := s.metaFallback.FallbackEnrich(quickCtx, input.URL)
	if !ok {
		return
	}

	s.applyContentEnrichment(input, gemini.NormalizeEnrichment(fallback))
}

func (s *bookmarkService) enrichAsync(userID, bookmarkID, rawURL string) {
	if s.enricher == nil && s.metaFallback == nil {
		return
	}

	go func() {
		defer func() {
			markCtx, markCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer markCancel()
			if err := s.repo.MarkEnriched(markCtx, userID, bookmarkID); err != nil && !errors.Is(err, domain.ErrBookmarkNotFound) {
				log.Printf("bookmark mark enriched failed for %s: %v", rawURL, err)
			}
		}()

		loadCtx, loadCancel := context.WithTimeout(context.Background(), 5*time.Second)
		hints, ok := s.loadEnrichmentHints(loadCtx, userID, bookmarkID)
		loadCancel()
		if !ok {
			return
		}

		// Stage 1: page-reading Enrich with retries, on its own time budget.
		hints, done := s.runEnrichLoop(userID, bookmarkID, rawURL, hints)
		if done {
			return
		}

		// Stage 2: always run the fallback + Classify with a FRESH budget, so
		// blocked or slow sites (e.g. HDRezka) still get a normalized title,
		// category and tags even when every Enrich attempt timed out.
		finalizeCtx, cancel := context.WithTimeout(context.Background(), enrichFinalizeTimeout)
		defer cancel()
		s.finalizeEnrichment(finalizeCtx, userID, bookmarkID, rawURL, hints)
	}()
}

// runEnrichLoop runs the page-reading Enrich step with retries. It returns the
// best hints found and whether the card is already good enough (so no fallback
// is needed). It uses its own time budget, separate from the Classify fallback.
func (s *bookmarkService) runEnrichLoop(userID, bookmarkID, rawURL string, hints domain.BookmarkEnrichment) (domain.BookmarkEnrichment, bool) {
	if s.enricher == nil {
		return hints, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), enrichLoopTimeout)
	defer cancel()

	for attempt := 1; attempt <= enrichMaxAttempts; attempt++ {
		attemptCtx, attemptCancel := context.WithTimeout(ctx, enrichAttemptTimeout)
		enrichment, err := s.enricher.Enrich(attemptCtx, rawURL)
		attemptCancel()

		if err == nil && !gemini.IsUnavailableEnrichment(enrichment) {
			finished := s.finishEnrichment(ctx, rawURL, hints, enrichment)
			imageURL := s.bookmarkImageURL(ctx, userID, bookmarkID)
			s.tryPersistEnrichment(ctx, userID, bookmarkID, rawURL, finished)
			if cardquality.IsGoodEnough(finished, imageURL) {
				return finished, true
			}
			hints = finished
		} else if err != nil {
			log.Printf("bookmark enrich attempt %d failed for %s: %v", attempt, rawURL, err)
		}

		if attempt == enrichMaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return hints, false
		case <-time.After(enrichRetryDelay(attempt)):
		}
	}

	return hints, false
}

func (s *bookmarkService) loadEnrichmentHints(ctx context.Context, userID, bookmarkID string) (domain.BookmarkEnrichment, bool) {
	bookmark, err := s.repo.GetByIDForUser(ctx, userID, bookmarkID)
	if err != nil {
		return domain.BookmarkEnrichment{}, false
	}
	return enrichmentFromBookmark(bookmark), true
}

func (s *bookmarkService) finishEnrichment(ctx context.Context, rawURL string, hints, enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	// Keep the rich classification from the main Enrich step (it reads the page
	// with the detailed prompt). Only fall back to Classify when it's still missing.
	merged := cardquality.Merge(hints, gemini.NormalizeEnrichment(enrichment))
	if classificationCompleteForURL(rawURL, merged) {
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
			merged = cardquality.Merge(merged, contentOnlyEnrichment(gemini.NormalizeEnrichment(fallback)))
		}
	}

	merged = s.classifyAndMerge(ctx, rawURL, merged)

	imageURL := s.bookmarkImageURL(ctx, userID, bookmarkID)
	s.tryPersistEnrichment(ctx, userID, bookmarkID, rawURL, merged)

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
	if s.enricher == nil || !hasContentForClassification(base) {
		return base
	}

	classifyCtx, cancel := context.WithTimeout(ctx, enrichAttemptTimeout)
	defer cancel()

	classified, err := s.enricher.Classify(classifyCtx, rawURL, base.Title, base.Description, titleSourceForClassification(rawURL, base.Title))
	if err != nil {
		log.Printf("bookmark classify failed for %s: %v", rawURL, err)
		return base
	}

	if gemini.IsUnavailableEnrichment(classified) {
		return base
	}

	return mergeClassifiedEnrichment(rawURL, base, classified)
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

func mergeClassifiedEnrichment(rawURL string, base, classified domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	normalized := gemini.NormalizeEnrichment(classified)
	merged := cardquality.Merge(base, normalized)

	source := titleSourceForClassification(rawURL, base.Title)
	if shouldPreserveBaseTitle(source, base.Title) {
		merged.Title = strings.TrimSpace(base.Title)
	} else if shouldPreferClassifiedTitle(source, base.Title, normalized.Title) {
		merged.Title = normalized.Title
	}

	return merged
}

func shouldPreserveBaseTitle(source, baseTitle string) bool {
	return source == "metadata_or_user" && cardquality.GoodTitle(baseTitle)
}

func shouldPreferClassifiedTitle(source, baseTitle, classifiedTitle string) bool {
	classifiedTitle = strings.TrimSpace(classifiedTitle)
	if classifiedTitle == "" || strings.EqualFold(strings.TrimSpace(baseTitle), classifiedTitle) {
		return false
	}

	return (source == "url_slug" || source == "url") && cardquality.GoodTitle(classifiedTitle)
}

func hasContentForClassification(enrichment domain.BookmarkEnrichment) bool {
	return strings.TrimSpace(enrichment.Title) != "" || strings.TrimSpace(enrichment.Description) != ""
}

// classificationComplete reports whether the card already has a confident
// category and tags, so the lighter Classify pass can be skipped.
func classificationComplete(enrichment domain.BookmarkEnrichment) bool {
	return enrichment.Category != "" && enrichment.Category != "other" && len(enrichment.Tags) >= 2
}

func classificationCompleteForURL(rawURL string, enrichment domain.BookmarkEnrichment) bool {
	if !classificationComplete(enrichment) {
		return false
	}

	// URL-derived titles can already have good category/tags but still need one
	// final Classify pass to normalize transliteration (e.g. "Garri Potter" or
	// "Klinki Hraniteley") without translating trusted metadata titles.
	source := titleSourceForClassification(rawURL, enrichment.Title)
	return source != "url_slug" && source != "url"
}

func titleSourceForClassification(rawURL, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "unknown"
	}

	if hint, ok := pagemeta.TitleHintFromURL(rawURL); ok && strings.EqualFold(title, strings.TrimSpace(hint)) {
		return "url_slug"
	}

	if strings.HasPrefix(title, "http://") || strings.HasPrefix(title, "https://") {
		return "url"
	}

	return "metadata_or_user"
}

func contentOnlyEnrichment(enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	return domain.BookmarkEnrichment{
		Title:       enrichment.Title,
		Description: enrichment.Description,
	}
}

func (s *bookmarkService) tryPersistEnrichment(ctx context.Context, userID, bookmarkID, rawURL string, enrichment domain.BookmarkEnrichment) bool {
	enrichment.Tags = pagemeta.EnsurePlatformTag(rawURL, enrichment.Tags)
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

func (s *bookmarkService) applyContentEnrichment(input *domain.CreateBookmarkInput, enrichment domain.BookmarkEnrichment) {
	content := contentOnlyEnrichment(enrichment)
	if input.Title == "" {
		input.Title = content.Title
	}
	if input.Description == "" {
		input.Description = content.Description
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
