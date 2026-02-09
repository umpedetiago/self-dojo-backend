package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrUsernameAlreadyExists = errors.New("username already exists")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, COALESCE(username, '') AS username, email, password_hash, created_at
FROM users
WHERE id = $1
`, id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, COALESCE(username, '') AS username, email, password_hash, created_at
FROM users
WHERE email = $1
`, email)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, username string, email string, passwordHash string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate uuid v7: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO users (id, username, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, password_hash, created_at
`, id, username, email, passwordHash)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return nil, ErrUsernameAlreadyExists
			}
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, username string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE users
SET username = $2
WHERE id = $1
RETURNING id, username, created_at, email
`, id, username)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.CreatedAt, &u.Email); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	result, err := r.pool.Exec(ctx, `
UPDATE users
SET password_hash = $2
WHERE id = $1
`, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
