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
	ID         int64       `json:"id"`
	Title      string      `json:"title"`
	SearchData string      `json:"search_data"`
	Link       string      `json:"link"`
	FileName   string      `json:"-"` // Internal use only, not exposed in API response
	Type       ContentType `json:"type"`
	Tags       []Tag       `json:"tags"`
	IsDeleted  bool        `json:"is_deleted"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// SearchFilter holds optional filters for the Search operation.
type SearchFilter struct {
	Query       string
	ContentType ContentType
	Tags        []string
}

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) ([]Content, error)
}

type ContentUsecase interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) ([]Content, error)
}
