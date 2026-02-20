package http

import (
	"time"

	"github.com/quoctann/content-hub/internal/domain"
)

// ContentResponse is the DTO exposed in API responses.
// It only contains fields safe to expose to clients.
type ContentResponse struct {
	ID        int64              `json:"id"`
	Title     *string            `json:"title"`
	TextData  *string            `json:"text_data"`
	Link      *string            `json:"link"`
	Type      domain.ContentType `json:"type"`
	Rank      float64            `json:"rank"`
	CreatedAt *time.Time         `json:"created_at"`
}

// ToContentResponse converts a domain.Content to ContentResponse.
func ToContentResponse(c *domain.Content) *ContentResponse {
	return &ContentResponse{
		ID:        c.ID,
		Title:     c.Title,
		TextData:  c.TextData,
		Link:      c.Link,
		Type:      c.Type,
		Rank:      c.Rank,
		CreatedAt: c.CreatedAt,
	}
}

// ToContentResponseList converts a slice of domain.Content to a slice of ContentResponse.
func ToContentResponseList(contents []domain.Content) []ContentResponse {
	responses := make([]ContentResponse, len(contents))
	for i, c := range contents {
		if resp := ToContentResponse(&c); resp != nil {
			responses[i] = *resp
		}
	}
	return responses
}

// PaginationMetaResponse is the DTO for pagination metadata in API responses.
type PaginationMetaResponse struct {
	TotalCount int64  `json:"total_count"`
	PageSize   int64  `json:"page_size"`
	Cursor     string `json:"cursor"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// ContentSearchResponseWrapper wraps search results with pagination metadata.
type ContentSearchResponseWrapper struct {
	Items      []ContentResponse      `json:"items"`
	Pagination PaginationMetaResponse `json:"pagination"`
}

// ToContentSearchResponse converts a domain.ContentSearchResult to ContentSearchResponseWrapper.
func ToContentSearchResponse(result *domain.ContentSearchResult) *ContentSearchResponseWrapper {
	return &ContentSearchResponseWrapper{
		Items: ToContentResponseList(result.Items),
		Pagination: PaginationMetaResponse{
			TotalCount: result.Pagination.TotalCount,
			PageSize:   result.Pagination.PageSize,
			Cursor:     result.Pagination.Cursor,
			NextCursor: result.Pagination.NextCursor,
			HasMore:    result.Pagination.HasMore,
		},
	}
}
