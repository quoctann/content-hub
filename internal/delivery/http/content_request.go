package http

import "github.com/quoctann/content-hub/internal/domain"

type CreateContentRequest struct {
	Title    *string `json:"title"`
	TextData *string `json:"text_data"`
	OCRText  *string `json:"ocr_text"`
	Caption  *string `json:"caption"`
	Link     *string `json:"link"`
	Type     string  `json:"type" binding:"required,oneof=text image video"`
}

// UpdateContentRequest is a partial update: omitted fields are left unchanged
// and an explicit null clears a nullable text field. type and is_hidden are
// not nullable, so for them null is treated the same as omitted.
type UpdateContentRequest struct {
	Title    NullableString `json:"title" swaggertype:"string"`
	TextData NullableString `json:"text_data" swaggertype:"string"`
	OCRText  NullableString `json:"ocr_text" swaggertype:"string"`
	Caption  NullableString `json:"caption" swaggertype:"string"`
	Link     NullableString `json:"link" swaggertype:"string"`
	Type     *string        `json:"type" binding:"omitempty,oneof=text image video"`
	IsHidden *bool          `json:"is_hidden"`
}

func (r UpdateContentRequest) ToPatch() domain.ContentPatch {
	patch := domain.ContentPatch{
		Title:    r.Title.toDomain(),
		TextData: r.TextData.toDomain(),
		OCRText:  r.OCRText.toDomain(),
		Caption:  r.Caption.toDomain(),
		Link:     r.Link.toDomain(),
	}
	if r.Type != nil {
		patch.Type = domain.Some(domain.ContentType(*r.Type))
	}
	if r.IsHidden != nil {
		patch.IsHidden = domain.Some(*r.IsHidden)
	}
	return patch
}

type BulkDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1,dive,min=1"`
}

type ToggleHideRequest struct {
	Hidden *bool `json:"hidden" binding:"required"`
}

type contentIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type SearchContentRequest struct {
	Keywords  string `form:"keywords"`
	MatchType string `form:"match_type" binding:"omitempty,oneof=and or"`
	Type      string `form:"type" binding:"omitempty,oneof=text image video"`
	Num       int64  `form:"num" binding:"omitempty,min=1,max=100"`
	Cursor    string `form:"cursor"`
}

type AdminListContentRequest struct {
	Page      int64  `form:"page" binding:"omitempty,min=1"`
	PageSize  int64  `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keywords  string `form:"keywords"`
	MatchType string `form:"match_type" binding:"omitempty,oneof=and or"`
	Type      string `form:"type" binding:"omitempty,oneof=text image video"`
	Visible   string `form:"visible" binding:"omitempty,oneof=true false"`
}

func (r CreateContentRequest) ToDomain() *domain.Content {
	content := &domain.Content{
		Title:    r.Title,
		TextData: r.TextData,
		OCRText:  r.OCRText,
		Caption:  r.Caption,
		Link:     r.Link,
	}
	if r.Type != "" {
		content.Type = domain.ContentType(r.Type)
	}
	return content
}
