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
	TextData  *string     `json:"text_data"` // Available for text content
	OCRText   *string     `json:"ocr_text"`  // Available for image content after OCR processing
	Caption   *string     `json:"caption"`   // Available for image content
	Link      *string     `json:"link"`
	FileName  *string     `json:"-"` // Internal use only
	Type      ContentType `json:"type"`
	IsHidden  bool        `json:"is_hidden"`
	Rank      float64     `json:"rank"` // Relevance rank from FTS, 0 if no search query
	CreatedAt *time.Time  `json:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at,omitempty"`
}

type SearchFilter struct {
	// Keywords is the preferred multi-term list. Each entry is an independent token;
	// MatchType controls whether all must appear (AND) or any suffices (OR).
	Keywords []string

	// MatchType is "and" or "or" (default "or"). Only meaningful when len(Keywords) > 1.
	MatchType string

	// ContentType restricts results to a specific media type ("text" or "image").
	// Empty means no restriction.
	ContentType ContentType

	// IncludeHidden controls whether hidden items can appear in results.
	//   false → hidden items are always excluded (public API default).
	//   true  → VisibilityFilter further refines which subset is returned (admin API).
	IncludeHidden bool

	// VisibilityFilter is only meaningful when IncludeHidden is true.
	//   "true"  → visible items only  (is_hidden = false)
	//   "false" → hidden items only   (is_hidden = true)
	//   ""      → no visibility restriction (all items)
	VisibilityFilter string
}

type PaginationMeta struct {
	TotalCount int64  `json:"total_count"`
	PageSize   int64  `json:"page_size"`
	Cursor     string `json:"cursor"`
	NextCursor string `json:"next_cursor"` // Empty string if no more results
	HasMore    bool   `json:"has_more"`
}

type ContentSearchResult struct {
	Items      []Content
	Pagination PaginationMeta
}

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Delete(ctx context.Context, id int64) error
	DeleteMany(ctx context.Context, ids []int64) error
	GetByID(ctx context.Context, id int64) (*Content, error)
	SetHidden(ctx context.Context, id int64, hidden bool) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) ([]Content, int64, error)
}

type ContentUsecase interface {
	Create(ctx context.Context, content *Content) error
	Update(ctx context.Context, content *Content) error
	Delete(ctx context.Context, id int64) error
	DeleteMany(ctx context.Context, ids []int64) error
	GetByID(ctx context.Context, id int64) (*Content, error)
	SetHidden(ctx context.Context, id int64, hidden bool) error
	Search(ctx context.Context, filter SearchFilter, cursor string, num int64) (*ContentSearchResult, error)
}
