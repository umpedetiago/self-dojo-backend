package config

import "os"

type Config struct {
	Port                     string
	DatabaseURL              string
	JWTSecret                string
	ResendAPIKey             string
	EmailFrom                string
	FrontendResetPasswordURL string
	SupabaseURL              string
	SupabaseServiceRoleKey   string
	AvatarBucket             string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/go_db?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		emailFrom = "onboarding@resend.dev"
	}

	avatarBucket := os.Getenv("AVATAR_BUCKET")
	if avatarBucket == "" {
		avatarBucket = "avatars"
	}

	return Config{
		Port:                     port,
		DatabaseURL:              dbURL,
		JWTSecret:                jwtSecret,
		ResendAPIKey:             os.Getenv("RESEND_API_KEY"),
		EmailFrom:                emailFrom,
		FrontendResetPasswordURL: os.Getenv("FRONTEND_RESET_PASSWORD_URL"),
		SupabaseURL:              os.Getenv("SUPABASE_URL"),
		SupabaseServiceRoleKey:   os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		AvatarBucket:             avatarBucket,
	}
}
