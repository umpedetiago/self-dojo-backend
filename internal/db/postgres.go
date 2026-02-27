package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return pool, nil
}

// ConnectWithRetry tenta conectar ao banco com retries (útil quando o Postgres ainda está subindo).
func ConnectWithRetry(ctx context.Context, databaseURL string, maxAttempts int, interval time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := Connect(ctx, databaseURL)
		if err == nil {
			return pool, nil
		}
		lastErr = err
		if attempt < maxAttempts {
			log.Printf("db: tentativa %d/%d falhou, aguardando %v: %v", attempt, maxAttempts, interval, err)
			time.Sleep(interval)
		}
	}
	return nil, fmt.Errorf("db connect after %d attempts: %w", maxAttempts, lastErr)
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, _ = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	var dataType string
	err := pool.QueryRow(ctx, `
SELECT data_type FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id'
`).Scan(&dataType)
	if err == nil && (dataType == "bigint" || dataType == "integer") {
		log.Printf("db: migrando users (id %s -> UUID), recriando tabela", dataType)
		_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS users`)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  display_name TEXT,
  photo_url TEXT,
  role TEXT NOT NULL DEFAULT 'student',
  martial_art_type TEXT,
  legacy_belt_id TEXT,
  legacy_degree INTEGER NOT NULL DEFAULT 0,
  legacy_total_classes INTEGER NOT NULL DEFAULT 0,
  legacy_has_aparadores BOOLEAN,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}
	// Migração: tabela já existente (UUID) sem username ganha a coluna
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT UNIQUE`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name TEXT`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS photo_url TEXT`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'student'`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS martial_art_type TEXT`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_belt_id TEXT`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_degree INTEGER NOT NULL DEFAULT 0`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_total_classes INTEGER NOT NULL DEFAULT 0`)
	_, _ = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS legacy_has_aparadores BOOLEAN`)

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS password_reset_tokens (
  token TEXT PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  used BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (password_reset_tokens): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS academies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  name TEXT NOT NULL,
  description TEXT,
  logo_url TEXT,
  address TEXT,
  city TEXT,
  state TEXT,
  phone TEXT,
  email TEXT,
  website TEXT,
  subscription_plan TEXT NOT NULL DEFAULT 'trial',
  subscription_started_at TIMESTAMPTZ,
  subscription_ends_at TIMESTAMPTZ,
  max_students INTEGER NOT NULL DEFAULT 10,
  max_teachers INTEGER NOT NULL DEFAULT 3,
  max_modalities INTEGER NOT NULL DEFAULT 3,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (academies): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS academy_modalities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academy_id UUID NOT NULL REFERENCES academies(id) ON DELETE CASCADE,
  martial_art_type TEXT NOT NULL,
  master_id UUID REFERENCES users(id) ON DELETE SET NULL,
  use_default_graduation BOOLEAN NOT NULL DEFAULT true,
  graduation_updated_at TIMESTAMPTZ,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(academy_id, martial_art_type)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (academy_modalities): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS belt_configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  modality_id UUID NOT NULL REFERENCES academy_modalities(id) ON DELETE CASCADE,
  belt_id TEXT NOT NULL,
  belt_name TEXT NOT NULL DEFAULT '',
  min_classes INTEGER NOT NULL DEFAULT 0,
  min_months INTEGER,
  min_classes_per_degree INTEGER,
  requires_exam BOOLEAN NOT NULL DEFAULT false,
  exam_fee DOUBLE PRECISION,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(modality_id, belt_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (belt_configs): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS academy_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academy_id UUID NOT NULL REFERENCES academies(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'student',
  status TEXT NOT NULL DEFAULT 'pending',
  joined_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(academy_id, user_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (academy_members): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS modality_teachers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  modality_id UUID NOT NULL REFERENCES academy_modalities(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'teacher',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(modality_id, user_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (modality_teachers): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS student_modalities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  member_id UUID NOT NULL REFERENCES academy_members(id) ON DELETE CASCADE,
  modality_id UUID NOT NULL REFERENCES academy_modalities(id) ON DELETE CASCADE,
  assigned_teacher_id UUID REFERENCES users(id) ON DELETE SET NULL,
  belt_id TEXT NOT NULL,
  degree INTEGER NOT NULL DEFAULT 0,
  promotion_date TIMESTAMPTZ,
  total_classes INTEGER NOT NULL DEFAULT 0,
  classes_at_current_belt INTEGER NOT NULL DEFAULT 0,
  enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, modality_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (student_modalities): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS graduation_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_modality_id UUID NOT NULL REFERENCES student_modalities(id) ON DELETE CASCADE,
  belt_id TEXT NOT NULL,
  degree INTEGER NOT NULL DEFAULT 0,
  promoted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  promoted_by UUID REFERENCES users(id) ON DELETE SET NULL,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (graduation_history): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS check_ins (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_modality_id UUID NOT NULL REFERENCES student_modalities(id) ON DELETE CASCADE,
  class_schedule_id UUID,
  checked_in_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  checked_in_by UUID REFERENCES users(id) ON DELETE SET NULL,
  class_type TEXT,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (check_ins): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS class_schedules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academy_id UUID NOT NULL REFERENCES academies(id) ON DELETE CASCADE,
  modality_id UUID REFERENCES academy_modalities(id) ON DELETE CASCADE,
  instructor_id UUID REFERENCES users(id) ON DELETE SET NULL,
  day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
  start_time TIME NOT NULL,
  end_time TIME NOT NULL CHECK (end_time > start_time),
  class_type TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  max_students INTEGER,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (class_schedules): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS student_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academy_id UUID NOT NULL REFERENCES academies(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(academy_id, name)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (student_groups): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS student_group_members (
  group_id UUID NOT NULL REFERENCES student_groups(id) ON DELETE CASCADE,
  academy_member_id UUID NOT NULL REFERENCES academy_members(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, academy_member_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (student_group_members): %w", err)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS class_schedule_groups (
  class_schedule_id UUID NOT NULL REFERENCES class_schedules(id) ON DELETE CASCADE,
  group_id UUID NOT NULL REFERENCES student_groups(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (class_schedule_id, group_id)
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema (class_schedule_groups): %w", err)
	}

	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_academies_owner ON academies(owner_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_academies_city ON academies(city)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_academy_modalities_academy ON academy_modalities(academy_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_academy_members_academy ON academy_members(academy_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_belt_configs_modality ON belt_configs(modality_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_student_modalities_member ON student_modalities(member_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_student_modalities_modality ON student_modalities(modality_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_graduation_history_student_modality ON graduation_history(student_modality_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_check_ins_student_modality ON check_ins(student_modality_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_check_ins_checked_in_at ON check_ins(checked_in_at)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_class_schedules_academy ON class_schedules(academy_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_class_schedules_modality ON class_schedules(modality_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_class_schedules_day ON class_schedules(day_of_week)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_class_schedules_active ON class_schedules(is_active)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_student_groups_academy ON student_groups(academy_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_student_group_members_member ON student_group_members(academy_member_id)`)
	_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_class_schedule_groups_group ON class_schedule_groups(group_id)`)
	_, _ = pool.Exec(ctx, `ALTER TABLE belt_configs ADD COLUMN IF NOT EXISTS belt_name TEXT NOT NULL DEFAULT ''`)
	return nil
}
