package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SupabaseConfig struct {
	BaseURL        string
	ServiceRoleKey string
	Bucket         string
}

type SupabaseProvider struct {
	baseURL        string
	serviceRoleKey string
	bucket         string
	client         *http.Client
}

func NewSupabaseProvider(cfg SupabaseConfig) (*SupabaseProvider, error) {
	baseURL := strings.TrimSpace(strings.TrimRight(cfg.BaseURL, "/"))
	if baseURL == "" {
		return nil, fmt.Errorf("supabase base url is required")
	}
	if strings.TrimSpace(cfg.ServiceRoleKey) == "" {
		return nil, fmt.Errorf("supabase service role key is required")
	}
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("supabase bucket is required")
	}

	return &SupabaseProvider{
		baseURL:        baseURL,
		serviceRoleKey: cfg.ServiceRoleKey,
		bucket:         bucket,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *SupabaseProvider) Upload(ctx context.Context, in UploadInput) (string, error) {
	path := strings.TrimLeft(strings.TrimSpace(in.Path), "/")
	if path == "" {
		return "", fmt.Errorf("upload path is required")
	}
	if len(in.Data) == 0 {
		return "", fmt.Errorf("upload payload is empty")
	}
	contentType := strings.TrimSpace(in.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(in.Data))
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload to supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("supabase upload failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	escapedPath := url.PathEscape(path)
	// PathEscape escapa "/" como %2F; para URL final precisamos manter separadores.
	escapedPath = strings.ReplaceAll(escapedPath, "%2F", "/")
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, escapedPath)
	return publicURL, nil
}

