package http

import (
	"time"

	"github.com/quoctann/content-hub/internal/domain"
)

type ContentResponse struct {
	ID        int64              `json:"id"`
	Title     *string            `json:"title"`
	TextData  *string            `json:"text_data"`
	OCRText   *string            `json:"ocr_text"`
	Caption   *string            `json:"caption"`
	Link      *string            `json:"link"`
	Type      domain.ContentType `json:"type"`
	Rank      float64            `json:"rank"`
	IsHidden  bool               `json:"is_hidden"`
	CreatedAt *time.Time         `json:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at"`
}

func ToContentResponse(c *domain.Content) *ContentResponse {
	return &ContentResponse{
		ID:        c.ID,
		Title:     c.Title,
		TextData:  c.TextData,
		OCRText:   c.OCRText,
		Caption:   c.Caption,
		Link:      c.Link,
		Type:      c.Type,
		Rank:      c.Rank,
		IsHidden:  c.IsHidden,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ToContentResponseList(contents []domain.Content) []ContentResponse {
	responses := make([]ContentResponse, len(contents))
	for i, c := range contents {
		if resp := ToContentResponse(&c); resp != nil {
			responses[i] = *resp
		}
	}
	return responses
}

type PaginationMetaResponse struct {
	TotalCount int64  `json:"total_count"`
	PageSize   int64  `json:"page_size"`
	Cursor     string `json:"cursor"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

type ContentSearchResponseWrapper struct {
	Items      []ContentResponse      `json:"items"`
	Pagination PaginationMetaResponse `json:"pagination"`
}

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

type AdminPaginationResponse struct {
	TotalCount int64 `json:"total_count"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}

type AdminContentResponseWrapper struct {
	Items      []ContentResponse       `json:"items"`
	Pagination AdminPaginationResponse `json:"pagination"`
}

func ToAdminContentResponse(result *domain.ContentSearchResult, page, pageSize int64) *AdminContentResponseWrapper {
	totalPages := (result.Pagination.TotalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	return &AdminContentResponseWrapper{
		Items: ToContentResponseList(result.Items),
		Pagination: AdminPaginationResponse{
			TotalCount: result.Pagination.TotalCount,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
}
