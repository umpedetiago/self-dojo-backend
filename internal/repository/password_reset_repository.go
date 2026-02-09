package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
var ErrTokenAlreadyUsed = errors.New("token already used")

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO password_reset_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
`, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) Get(ctx context.Context, token string) (*PasswordResetToken, error) {
	row := r.pool.QueryRow(ctx, `
SELECT token, user_id, expires_at, used, created_at
FROM password_reset_tokens
WHERE token = $1
`, token)

	var t PasswordResetToken
	if err := row.Scan(&t.Token, &t.UserID, &t.ExpiresAt, &t.Used, &t.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("get password reset token: %w", err)
	}

	if t.Used {
		return nil, ErrTokenAlreadyUsed
	}

	if time.Now().After(t.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return &t, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE password_reset_tokens
SET used = true
WHERE token = $1
`, token)
	if err != nil {
		return fmt.Errorf("mark token as used: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
DELETE FROM password_reset_tokens
WHERE expires_at < now() OR used = true
`)
	if err != nil {
		return fmt.Errorf("cleanup expired tokens: %w", err)
	}
	return nil
}
