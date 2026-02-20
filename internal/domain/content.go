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
	Title     *string     `json:"title"`
	TextData  *string     `json:"text_data"` // available for text content
	OCRText   *string     `json:"ocr_text"`  // available for image content after OCR processing
	Caption   *string     `json:"caption"`   // available for image content
	Link      *string     `json:"link"`
	FileName  *string     `json:"-"` // Internal use only, not exposed in API response
	Type      ContentType `json:"type"`
	Rank      float64     `json:"rank"` // Relevance rank from FTS, 0 if no search query
	CreatedAt *time.Time  `json:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at"`
}

// SearchFilter holds optional filters for the Search operation.
type SearchFilter struct {
	Query       string
	Keywords    []string
	MatchType   string // "and" or "or", default "or"
	ContentType ContentType
}

// PaginationMeta holds pagination metadata for paginated responses
type PaginationMeta struct {
	TotalCount int64  `json:"total_count"`
	PageSize   int64  `json:"page_size"`
	Cursor     string `json:"cursor"`
	NextCursor string `json:"next_cursor"` // Empty string if no more results
	HasMore    bool   `json:"has_more"`
}

// ContentSearchResult wraps search results with pagination metadata
type ContentSearchResult struct {
	Items      []Content
	Pagination PaginationMeta
}

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) ([]Content, int64, error)
}

type ContentUsecase interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) (*ContentSearchResult, error)
}
