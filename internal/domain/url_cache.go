package domain

import "time"

type URLEnrichmentCacheEntry struct {
	CanonicalURL string
	Title        string
	Description  string
	Category     string
	Tags         []string
	ImageURL     string
	EnrichedAt   time.Time
}

func (e URLEnrichmentCacheEntry) Enrichment() BookmarkEnrichment {
	return BookmarkEnrichment{
		Title:       e.Title,
		Description: e.Description,
		Category:    e.Category,
		Tags:        e.Tags,
	}
}
