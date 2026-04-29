package http

type UpdateContentRequest struct {
	Title    *string `json:"title"`
	TextData *string `json:"text_data"`
	OCRText  *string `json:"ocr_text"`
	Caption  *string `json:"caption"`
	Link     *string `json:"link"`
	Type     *string `json:"type"`
	IsHidden *bool   `json:"is_hidden"`
}

type BulkDeleteRequest struct {
	IDs []int64 `json:"ids"`
}
