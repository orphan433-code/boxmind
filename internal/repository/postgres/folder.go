package postgres

import (
	"context"
	"errors"
	"fmt"

	"pet-link/internal/domain"

	"github.com/jackc/pgx/v5"
)

type FolderRepository struct {
	db *DB
}

func NewFolderRepository(db *DB) *FolderRepository {
	return &FolderRepository{db: db}
}

func (r *FolderRepository) Create(ctx context.Context, userID string, input domain.CreateFolderInput) (domain.Folder, error) {
	const query = `
		INSERT INTO bookmark_folders (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at, updated_at
	`

	var folder domain.Folder
	err := r.db.Pool.QueryRow(ctx, query, userID, input.Name).Scan(
		&folder.ID, &folder.UserID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Folder{}, domain.ErrFolderAlreadyExists
		}
		return domain.Folder{}, fmt.Errorf("create folder: %w", err)
	}

	return folder, nil
}

func (r *FolderRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Folder, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM bookmark_folders
		WHERE user_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	defer rows.Close()

	folders := make([]domain.Folder, 0)
	for rows.Next() {
		var folder domain.Folder
		if err := rows.Scan(&folder.ID, &folder.UserID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate folders: %w", err)
	}

	return folders, nil
}

func (r *FolderRepository) GetByIDForUser(ctx context.Context, userID, folderID string) (domain.Folder, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM bookmark_folders
		WHERE id = $1 AND user_id = $2
	`

	var folder domain.Folder
	err := r.db.Pool.QueryRow(ctx, query, folderID, userID).Scan(
		&folder.ID, &folder.UserID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Folder{}, domain.ErrFolderNotFound
		}
		return domain.Folder{}, fmt.Errorf("get folder: %w", err)
	}

	return folder, nil
}

func (r *FolderRepository) UpdateName(ctx context.Context, userID, folderID, name string) (domain.Folder, error) {
	const query = `
		UPDATE bookmark_folders
		SET name = $1, updated_at = now()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, name, created_at, updated_at
	`

	var folder domain.Folder
	err := r.db.Pool.QueryRow(ctx, query, name, folderID, userID).Scan(
		&folder.ID, &folder.UserID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Folder{}, domain.ErrFolderNotFound
		}
		if isUniqueViolation(err) {
			return domain.Folder{}, domain.ErrFolderAlreadyExists
		}
		return domain.Folder{}, fmt.Errorf("update folder: %w", err)
	}

	return folder, nil
}

func (r *FolderRepository) Delete(ctx context.Context, userID, folderID string) error {
	const query = `
		DELETE FROM bookmark_folders
		WHERE id = $1 AND user_id = $2
	`

	tag, err := r.db.Pool.Exec(ctx, query, folderID, userID)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFolderNotFound
	}

	return nil
}
