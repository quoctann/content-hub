package domain

import "context"

type Article struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleRepository interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]Article, error)
	GetByID(ctx context.Context, id int64) (Article, error)
	Store(ctx context.Context, a *Article) error
}

type ArticleUsecase interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]Article, error)
	GetByID(ctx context.Context, id int64) (Article, error)
	Store(ctx context.Context, a *Article) error
}
