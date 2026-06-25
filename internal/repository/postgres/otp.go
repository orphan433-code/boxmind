package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pet-link/internal/domain"

	"github.com/jackc/pgx/v5"
)

type OTPRepository struct {
	db *DB
}

func NewOTPRepository(db *DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) InvalidateActiveByEmail(ctx context.Context, email string) error {
	const query = `
		UPDATE login_otps
		SET used_at = now()
		WHERE email = $1 AND used_at IS NULL
	`

	if _, err := r.db.Pool.Exec(ctx, query, email); err != nil {
		return fmt.Errorf("invalidate active otps: %w", err)
	}

	return nil
}

func (r *OTPRepository) Create(ctx context.Context, email, codeHash string, expiresAt time.Time) error {
	if err := r.InvalidateActiveByEmail(ctx, email); err != nil {
		return err
	}

	const query = `
		INSERT INTO login_otps (email, code_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	if _, err := r.db.Pool.Exec(ctx, query, email, codeHash, expiresAt); err != nil {
		return fmt.Errorf("create login otp: %w", err)
	}

	return nil
}

func (r *OTPRepository) GetLatestActive(ctx context.Context, email string) (domain.LoginOTP, error) {
	const query = `
		SELECT id, email, code_hash, expires_at, used_at, created_at
		FROM login_otps
		WHERE email = $1 AND used_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp domain.LoginOTP
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&otp.ID,
		&otp.Email,
		&otp.CodeHash,
		&otp.ExpiresAt,
		&otp.UsedAt,
		&otp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LoginOTP{}, domain.ErrInvalidOTP
		}
		return domain.LoginOTP{}, fmt.Errorf("get latest active otp: %w", err)
	}

	return otp, nil
}

func (r *OTPRepository) MarkUsed(ctx context.Context, id string) error {
	const query = `
		UPDATE login_otps
		SET used_at = now()
		WHERE id = $1 AND used_at IS NULL
	`

	tag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark otp used: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrInvalidOTP
	}

	return nil
}
