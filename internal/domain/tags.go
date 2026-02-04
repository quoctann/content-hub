package domain

import (
	"context"
	"time"
)

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	Search(ctx context.Context, query string, cursor string, num int64) ([]Tag, error)
}

type TagUsecase interface {
	Create(ctx context.Context, tag *Tag) error
	Search(ctx context.Context, query string, cursor string, num int64) ([]Tag, error)
}
