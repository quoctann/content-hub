package domain

import (
	"context"
	"time"
)

type ContentType string

const (
	Text  ContentType = "text"
	Image ContentType = "image"
)

type Content struct {
	ID        int64       `json:"id"`
	Title     string      `json:"title"`
	Text      string      `json:"text"`
	URL       string      `json:"url"`
	Type      ContentType `json:"type"`
	Tags      []Tag       `json:"tags"`
	IsDeleted bool        `json:"is_deleted"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, query string, cursor string, num int64) ([]Content, error)
}

type ContentUsecase interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, query string, cursor string, num int64) ([]Content, error)
}
