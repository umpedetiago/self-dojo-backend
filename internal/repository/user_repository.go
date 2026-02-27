package repository

import (
	"context"
	"database/sql"
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
	DisplayName  string
	PhotoURL     string
	Role         string
	MartialArt   string
	LegacyBeltID string
	LegacyDegree int
	LegacyTotal  int
	HasAparador  *bool
	CreatedAt    time.Time
}

type UpdateUserProfileInput struct {
	Username           *string
	DisplayName        *string
	PhotoURL           *string
	Role               *string
	MartialArtType     *string
	LegacyBeltID       *string
	LegacyDegree       *int
	LegacyTotalClasses *int
	LegacyHasAparador  *bool
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
SELECT
  id,
  COALESCE(username, '') AS username,
  email,
  password_hash,
  COALESCE(display_name, '') AS display_name,
  COALESCE(photo_url, '') AS photo_url,
  COALESCE(role, 'student') AS role,
  COALESCE(martial_art_type, '') AS martial_art_type,
  COALESCE(legacy_belt_id, '') AS legacy_belt_id,
  COALESCE(legacy_degree, 0) AS legacy_degree,
  COALESCE(legacy_total_classes, 0) AS legacy_total_classes,
  legacy_has_aparadores,
  created_at
FROM users
WHERE id = $1
`, id)

	var u User
	var hasAparadores sql.NullBool
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.PhotoURL,
		&u.Role,
		&u.MartialArt,
		&u.LegacyBeltID,
		&u.LegacyDegree,
		&u.LegacyTotal,
		&hasAparadores,
		&u.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if hasAparadores.Valid {
		v := hasAparadores.Bool
		u.HasAparador = &v
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  id,
  COALESCE(username, '') AS username,
  email,
  password_hash,
  COALESCE(display_name, '') AS display_name,
  COALESCE(photo_url, '') AS photo_url,
  COALESCE(role, 'student') AS role,
  COALESCE(martial_art_type, '') AS martial_art_type,
  COALESCE(legacy_belt_id, '') AS legacy_belt_id,
  COALESCE(legacy_degree, 0) AS legacy_degree,
  COALESCE(legacy_total_classes, 0) AS legacy_total_classes,
  legacy_has_aparadores,
  created_at
FROM users
WHERE email = $1
`, email)

	var u User
	var hasAparadores sql.NullBool
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.PhotoURL,
		&u.Role,
		&u.MartialArt,
		&u.LegacyBeltID,
		&u.LegacyDegree,
		&u.LegacyTotal,
		&hasAparadores,
		&u.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if hasAparadores.Valid {
		v := hasAparadores.Bool
		u.HasAparador = &v
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
RETURNING
  id,
  username,
  email,
  password_hash,
  COALESCE(display_name, '') AS display_name,
  COALESCE(photo_url, '') AS photo_url,
  COALESCE(role, 'student') AS role,
  COALESCE(martial_art_type, '') AS martial_art_type,
  COALESCE(legacy_belt_id, '') AS legacy_belt_id,
  COALESCE(legacy_degree, 0) AS legacy_degree,
  COALESCE(legacy_total_classes, 0) AS legacy_total_classes,
  legacy_has_aparadores,
  created_at
`, id, username, email, passwordHash)

	var u User
	var hasAparadores sql.NullBool
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.PhotoURL,
		&u.Role,
		&u.MartialArt,
		&u.LegacyBeltID,
		&u.LegacyDegree,
		&u.LegacyTotal,
		&hasAparadores,
		&u.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return nil, ErrUsernameAlreadyExists
			}
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	if hasAparadores.Valid {
		v := hasAparadores.Bool
		u.HasAparador = &v
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, username string) (*User, error) {
	return r.UpdateProfile(ctx, id, UpdateUserProfileInput{
		Username: &username,
	})
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateUserProfileInput) (*User, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE users
SET
  username = COALESCE($2, username),
  display_name = COALESCE($3, display_name),
  photo_url = COALESCE($4, photo_url),
  role = COALESCE($5, role),
  martial_art_type = COALESCE($6, martial_art_type),
  legacy_belt_id = COALESCE($7, legacy_belt_id),
  legacy_degree = COALESCE($8, legacy_degree),
  legacy_total_classes = COALESCE($9, legacy_total_classes),
  legacy_has_aparadores = COALESCE($10, legacy_has_aparadores)
WHERE id = $1
RETURNING
  id,
  COALESCE(username, '') AS username,
  email,
  password_hash,
  COALESCE(display_name, '') AS display_name,
  COALESCE(photo_url, '') AS photo_url,
  COALESCE(role, 'student') AS role,
  COALESCE(martial_art_type, '') AS martial_art_type,
  COALESCE(legacy_belt_id, '') AS legacy_belt_id,
  COALESCE(legacy_degree, 0) AS legacy_degree,
  COALESCE(legacy_total_classes, 0) AS legacy_total_classes,
  legacy_has_aparadores,
  created_at
`, id, in.Username, in.DisplayName, in.PhotoURL, in.Role, in.MartialArtType, in.LegacyBeltID, in.LegacyDegree, in.LegacyTotalClasses, in.LegacyHasAparador)

	var u User
	var hasAparadores sql.NullBool
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.PhotoURL,
		&u.Role,
		&u.MartialArt,
		&u.LegacyBeltID,
		&u.LegacyDegree,
		&u.LegacyTotal,
		&hasAparadores,
		&u.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	if hasAparadores.Valid {
		v := hasAparadores.Bool
		u.HasAparador = &v
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
