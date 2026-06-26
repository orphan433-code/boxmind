package service

import (
	"context"
	"log"
	"strings"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/cardquality"
	"pet-link/internal/pkg/gemini"
	"pet-link/internal/pkg/pagemeta"
)

type URLEnrichmentCacheRepository interface {
	Get(ctx context.Context, canonicalURL string) (domain.URLEnrichmentCacheEntry, bool, error)
	Upsert(ctx context.Context, entry domain.URLEnrichmentCacheEntry) error
}

func (s *bookmarkService) cacheGet(ctx context.Context, rawURL string) (domain.URLEnrichmentCacheEntry, bool, error) {
	if s.cache == nil {
		return domain.URLEnrichmentCacheEntry{}, false, nil
	}
	return s.cache.Get(ctx, rawURL)
}

func isUsableCacheEntry(entry domain.URLEnrichmentCacheEntry) bool {
	enrichment := entry.Enrichment()
	if gemini.IsUnavailableEnrichment(enrichment) {
		return false
	}
	return cardquality.IsAcceptable(enrichment, entry.ImageURL) && !cardquality.NeedsPolish(enrichment, entry.ImageURL)
}

func applyCacheToCreateInput(input *domain.CreateBookmarkInput, entry domain.URLEnrichmentCacheEntry) {
	if input.Title == "" && strings.TrimSpace(entry.Title) != "" {
		input.Title = entry.Title
	}
	if input.Description == "" && strings.TrimSpace(entry.Description) != "" {
		input.Description = entry.Description
	}
	if (input.Category == "" || input.Category == "other") &&
		entry.Category != "" && entry.Category != "other" {
		input.Category = entry.Category
	}
	if len(input.Tags) < 2 && len(entry.Tags) >= 2 {
		input.Tags = append([]string(nil), entry.Tags...)
	}
	if input.ImageURL == "" && strings.TrimSpace(entry.ImageURL) != "" {
		input.ImageURL = entry.ImageURL
	}
}

func (s *bookmarkService) prefetchCreateCache(ctx context.Context, input *domain.CreateBookmarkInput) {
	if s.cache == nil {
		return
	}

	entry, ok, err := s.cache.Get(ctx, input.URL)
	if err != nil {
		log.Printf("[ENRICH-CACHE] lookup failed url=%s err=%v", input.URL, err)
		return
	}
	if !ok || !isUsableCacheEntry(entry) {
		return
	}

	applyCacheToCreateInput(input, entry)
	log.Printf("[ENRICH-CACHE] prefetch url=%s category=%s tags=%d", input.URL, entry.Category, len(entry.Tags))
}

func (s *bookmarkService) applyCachedEnrichment(ctx context.Context, userID, bookmarkID, rawURL string, entry domain.URLEnrichmentCacheEntry) bool {
	enrichment := entry.Enrichment()
	enrichment.Tags = pagemeta.EnsurePlatformTag(rawURL, enrichment.Tags)

	if err := s.persistEnrichment(ctx, userID, bookmarkID, enrichment); err != nil {
		log.Printf("[ENRICH-CACHE] apply failed url=%s err=%v", rawURL, err)
		return false
	}

	if entry.ImageURL != "" {
		bookmark, err := s.repo.GetByIDForUser(ctx, userID, bookmarkID)
		if err == nil && strings.TrimSpace(bookmark.ImageURL) == "" {
			if err := s.repo.UpdateImageURL(ctx, userID, bookmarkID, entry.ImageURL); err != nil {
				log.Printf("[ENRICH-CACHE] image update failed url=%s err=%v", rawURL, err)
			}
		}
	}

	log.Printf("[ENRICH-CACHE] hit url=%s category=%s tags=%d", rawURL, entry.Category, len(entry.Tags))
	return true
}

func (s *bookmarkService) storeEnrichmentCache(ctx context.Context, rawURL string, enrichment domain.BookmarkEnrichment, imageURL string) {
	if s.cache == nil {
		return
	}
	if strings.TrimSpace(imageURL) == "" {
		imageURL = pagemeta.PlatformThumbnailURL(rawURL)
	}
	if !cardquality.IsAcceptable(enrichment, imageURL) {
		return
	}

	entry := domain.URLEnrichmentCacheEntry{
		CanonicalURL: rawURL,
		Title:        enrichment.Title,
		Description:  enrichment.Description,
		Category:     enrichment.Category,
		Tags:         append([]string(nil), enrichment.Tags...),
		ImageURL:     imageURL,
	}

	if err := s.cache.Upsert(ctx, entry); err != nil {
		log.Printf("[ENRICH-CACHE] store failed url=%s err=%v", rawURL, err)
		return
	}

	log.Printf("[ENRICH-CACHE] store url=%s category=%s tags=%d", rawURL, entry.Category, len(entry.Tags))
}
