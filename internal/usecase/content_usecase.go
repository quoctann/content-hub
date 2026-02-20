package usecase

import (
	"context"
	"fmt"
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

func (u *contentUsecase) Search(ctx context.Context, filter domain.SearchFilter, cursor string, num int64) (*domain.ContentSearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Validate num
	if num <= 0 {
		num = 10
	}

	// Get items and total count from repository
	items, total, err := u.contentRepo.Search(ctx, filter, cursor, num)
	if err != nil {
		return nil, err
	}

	// Parse cursor as offset
	var offset int64 = 0
	if cursor != "" {
		_, _ = fmt.Sscanf(cursor, "%d", &offset)
	}

	// Calculate next cursor
	nextCursor := ""
	hasMore := false
	if int64(len(items)) == num && offset+num < total {
		hasMore = true
		nextCursor = fmt.Sprintf("%d", offset+num)
	}

	return &domain.ContentSearchResult{
		Items: items,
		Pagination: domain.PaginationMeta{
			TotalCount: total,
			PageSize:   num,
			Cursor:     cursor,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}, nil
}
