package service

import (
	"context"
	"log"
	"strings"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/cardquality"
)

// eligibleForClassifyFirst reports whether we already have enough text to try
// the cheap Classify pass before the expensive URL-context Enrich step.
func eligibleForClassifyFirst(rawURL string, enrichment domain.BookmarkEnrichment) bool {
	if !hasContentForClassification(enrichment) {
		return false
	}

	title := strings.TrimSpace(enrichment.Title)
	if title == "" {
		return false
	}

	switch titleSourceForClassification(rawURL, title) {
	case "url":
		return false
	case "url_slug", "metadata_or_user":
		return cardquality.GoodTitle(title)
	default:
		return cardquality.GoodTitle(title)
	}
}

func (s *bookmarkService) runClassifyFirst(
	ctx context.Context,
	userID, bookmarkID, rawURL string,
	hints domain.BookmarkEnrichment,
) (domain.BookmarkEnrichment, bool) {
	if s.enricher == nil || !eligibleForClassifyFirst(rawURL, hints) {
		return hints, false
	}

	merged := s.classifyAndMerge(ctx, rawURL, hints)
	imageURL := s.bookmarkImageURL(ctx, userID, bookmarkID)
	merged = s.applyMovieMetadataIfNeeded(ctx, userID, bookmarkID, rawURL, merged)
	imageURL = s.bookmarkImageURL(ctx, userID, bookmarkID)
	merged = s.applyTextPolishIfNeeded(ctx, rawURL, merged, imageURL)
	s.tryPersistEnrichment(ctx, userID, bookmarkID, rawURL, merged)

	if !cardquality.IsAcceptable(merged, imageURL) {
		log.Printf(
			"[ENRICH-CLASSIFY-FIRST] insufficient url=%s category=%s tags=%d",
			rawURL,
			merged.Category,
			len(merged.Tags),
		)
		return merged, false
	}

	log.Printf(
		"[ENRICH-CLASSIFY-FIRST] success url=%s category=%s tags=%d",
		rawURL,
		merged.Category,
		len(merged.Tags),
	)
	return merged, true
}
