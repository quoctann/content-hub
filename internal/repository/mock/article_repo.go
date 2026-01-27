package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/quoctann/content-hub/internal/domain"
)

type mockArticleRepo struct {
	articles map[int64]domain.Article
	mu       sync.RWMutex
	lastID   int64
}

func NewMockArticleRepo() domain.ArticleRepository {
	return &mockArticleRepo{
		articles: make(map[int64]domain.Article),
	}
}

func (m *mockArticleRepo) Fetch(ctx context.Context, cursor string, num int64) ([]domain.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.Article, 0, len(m.articles))
	for _, a := range m.articles {
		res = append(res, a)
	}
	return res, nil
}

func (m *mockArticleRepo) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.articles[id]; ok {
		return a, nil
	}
	return domain.Article{}, errors.New("article not found")
}

func (m *mockArticleRepo) Store(ctx context.Context, a *domain.Article) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastID++
	a.ID = m.lastID
	m.articles[a.ID] = *a
	return nil
}
