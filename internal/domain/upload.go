package domain

import (
	"context"
	"io"
)

type UploadInput struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

type UploadResult struct {
	URL string
}

// MediaStore isolates content management from a particular object-storage provider.
type MediaStore interface {
	Upload(ctx context.Context, input UploadInput) (UploadResult, error)
}
