package postgres

import (
	"context"
	"errors"
	"fmt"

	"pet-link/internal/domain"

	"github.com/jackc/pgx/v5"
)

type URLEnrichmentCacheRepository struct {
	db *DB
}

func NewURLEnrichmentCacheRepository(db *DB) *URLEnrichmentCacheRepository {
	return &URLEnrichmentCacheRepository{db: db}
}

func (r *URLEnrichmentCacheRepository) Get(ctx context.Context, canonicalURL string) (domain.URLEnrichmentCacheEntry, bool, error) {
	const query = `
		SELECT canonical_url, title, description, category, tags, image_url, enriched_at
		FROM url_enrichment_cache
		WHERE canonical_url = $1
	`

	var entry domain.URLEnrichmentCacheEntry
	err := r.db.Pool.QueryRow(ctx, query, canonicalURL).Scan(
		&entry.CanonicalURL,
		&entry.Title,
		&entry.Description,
		&entry.Category,
		&entry.Tags,
		&entry.ImageURL,
		&entry.EnrichedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.URLEnrichmentCacheEntry{}, false, nil
		}
		return domain.URLEnrichmentCacheEntry{}, false, fmt.Errorf("get url enrichment cache: %w", err)
	}

	return entry, true, nil
}

func (r *URLEnrichmentCacheRepository) Upsert(ctx context.Context, entry domain.URLEnrichmentCacheEntry) error {
	const query = `
		INSERT INTO url_enrichment_cache (
			canonical_url, title, description, category, tags, image_url, enriched_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (canonical_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			image_url = CASE WHEN EXCLUDED.image_url <> '' THEN EXCLUDED.image_url ELSE url_enrichment_cache.image_url END,
			enriched_at = now(),
			updated_at = now()
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		entry.CanonicalURL,
		entry.Title,
		entry.Description,
		entry.Category,
		entry.Tags,
		entry.ImageURL,
	)
	if err != nil {
		return fmt.Errorf("upsert url enrichment cache: %w", err)
	}

	return nil
}
