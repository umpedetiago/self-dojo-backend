package storage

import "context"

type UploadInput struct {
	Path        string
	ContentType string
	Data        []byte
}

type Provider interface {
	Upload(ctx context.Context, in UploadInput) (publicURL string, err error)
}

