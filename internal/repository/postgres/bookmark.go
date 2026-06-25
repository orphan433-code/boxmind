package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/bookmarkurl"

	"github.com/jackc/pgx/v5"
)

type BookmarkRepository struct {
	db *DB
}

func NewBookmarkRepository(db *DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) Create(ctx context.Context, userID string, input domain.CreateBookmarkInput) (domain.Bookmark, error) {
	const query = `
		INSERT INTO bookmarks (user_id, url, title, description, image_url, category, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, url, title, description, image_url, category, tags, created_at, updated_at
	`

	var b domain.Bookmark
	err := r.db.Pool.QueryRow(ctx, query, userID, input.URL, input.Title, input.Description, input.ImageURL, input.Category, input.Tags).Scan(
		&b.ID, &b.UserID, &b.URL, &b.Title, &b.Description, &b.ImageURL, &b.Category, &b.Tags, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Bookmark{}, domain.ErrBookmarkAlreadyExists
		}
		return domain.Bookmark{}, fmt.Errorf("create bookmark: %w", err)
	}

	return b, nil
}

func (r *BookmarkRepository) UpdateImageURL(ctx context.Context, userID, bookmarkID, imageURL string) error {
	const query = `
		UPDATE bookmarks
		SET image_url = $1, updated_at = now()
		WHERE id = $2 AND user_id = $3
	`

	tag, err := r.db.Pool.Exec(ctx, query, imageURL, bookmarkID, userID)
	if err != nil {
		return fmt.Errorf("update bookmark image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookmarkNotFound
	}

	return nil
}

func (r *BookmarkRepository) UpdateEnrichment(ctx context.Context, userID, bookmarkID string, enrichment domain.BookmarkEnrichment) error {
	const query = `
		UPDATE bookmarks
		SET
			title = CASE WHEN $1 <> '' THEN $1 ELSE title END,
			description = CASE WHEN $2 <> '' THEN $2 ELSE description END,
			category = CASE WHEN $3 <> '' THEN $3 ELSE category END,
			tags = CASE WHEN cardinality($4::text[]) > 0 THEN $4 ELSE tags END,
			updated_at = now()
		WHERE id = $5 AND user_id = $6
	`

	tag, err := r.db.Pool.Exec(ctx, query, enrichment.Title, enrichment.Description, enrichment.Category, enrichment.Tags, bookmarkID, userID)
	if err != nil {
		return fmt.Errorf("update bookmark enrichment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookmarkNotFound
	}

	return nil
}

func (r *BookmarkRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Bookmark, error) {
	const query = `
		SELECT id, user_id, url, title, description, image_url, category, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]domain.Bookmark, 0)
	for rows.Next() {
		var b domain.Bookmark
		if err := rows.Scan(&b.ID, &b.UserID, &b.URL, &b.Title, &b.Description, &b.ImageURL, &b.Category, &b.Tags, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookmarks: %w", err)
	}

	return bookmarks, nil
}

func (r *BookmarkRepository) GetByIDForUser(ctx context.Context, userID, bookmarkID string) (domain.Bookmark, error) {
	const query = `
		SELECT id, user_id, url, title, description, image_url, category, tags, created_at, updated_at
		FROM bookmarks
		WHERE id = $1 AND user_id = $2
	`

	var b domain.Bookmark
	err := r.db.Pool.QueryRow(ctx, query, bookmarkID, userID).Scan(
		&b.ID, &b.UserID, &b.URL, &b.Title, &b.Description, &b.ImageURL, &b.Category, &b.Tags, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Bookmark{}, domain.ErrBookmarkNotFound
		}
		return domain.Bookmark{}, fmt.Errorf("get bookmark: %w", err)
	}

	return b, nil
}

func (r *BookmarkRepository) ExistsByURLForUser(ctx context.Context, userID, canonicalURL string) (bool, error) {
	const query = `SELECT url FROM bookmarks WHERE user_id = $1`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return false, fmt.Errorf("list bookmark urls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stored string
		if err := rows.Scan(&stored); err != nil {
			return false, fmt.Errorf("scan bookmark url: %w", err)
		}

		normalized, err := bookmarkurl.Normalize(stored)
		if err != nil {
			if strings.TrimSpace(stored) == canonicalURL {
				return true, nil
			}
			continue
		}
		if normalized == canonicalURL {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate bookmark urls: %w", err)
	}

	return false, nil
}

func (r *BookmarkRepository) Delete(ctx context.Context, userID, bookmarkID string) error {
	const query = `
		DELETE FROM bookmarks
		WHERE id = $1 AND user_id = $2
	`

	tag, err := r.db.Pool.Exec(ctx, query, bookmarkID, userID)
	if err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrBookmarkNotFound
	}

	return nil
}
