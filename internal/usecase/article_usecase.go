package usecase

import (
	"context"
	"time"

	"github.com/quoctann/content-hub/internal/domain"
)

type articleUsecase struct {
	articleRepo    domain.ArticleRepository
	contextTimeout time.Duration
}

func NewArticleUsecase(a domain.ArticleRepository, timeout time.Duration) domain.ArticleUsecase {
	return &articleUsecase{
		articleRepo:    a,
		contextTimeout: timeout,
	}
}

func (a *articleUsecase) Fetch(ctx context.Context, cursor string, num int64) ([]domain.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, a.contextTimeout)
	defer cancel()
	return a.articleRepo.Fetch(ctx, cursor, num)
}

func (a *articleUsecase) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, a.contextTimeout)
	defer cancel()
	return a.articleRepo.GetByID(ctx, id)
}

func (a *articleUsecase) Store(ctx context.Context, article *domain.Article) error {
	ctx, cancel := context.WithTimeout(ctx, a.contextTimeout)
	defer cancel()
	return a.articleRepo.Store(ctx, article)
}
