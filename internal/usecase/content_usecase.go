package usecase

import (
	"context"
	"time"

	"github.com/quoctann/content-hub/internal/domain"
)

type contentUsecase struct {
	contentRepo    domain.ContentRepository
	contextTimeout time.Duration
}

func NewContentUsecase(c domain.ContentRepository, timeout time.Duration) domain.ContentUsecase {
	return &contentUsecase{
		contentRepo:    c,
		contextTimeout: timeout,
	}
}

func (u *contentUsecase) Create(ctx context.Context, content *domain.Content) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.contentRepo.Create(ctx, content)
}

func (u *contentUsecase) Update(ctx context.Context, content *domain.Content) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.contentRepo.Update(ctx, content)
}

func (u *contentUsecase) Search(ctx context.Context, query string, cursor string, num int64) ([]domain.Content, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.contentRepo.Search(ctx, query, cursor, num)
}
