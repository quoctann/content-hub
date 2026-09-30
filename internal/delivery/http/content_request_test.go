package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func bindUpdate(t *testing.T, body string) UpdateContentRequest {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/contents/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req UpdateContentRequest
	if err := c.ShouldBind(&req); err != nil {
		t.Fatalf("ShouldBind(%s) error = %v", body, err)
	}
	return req
}

func TestUpdateContentRequestNullSemantics(t *testing.T) {
	patch := bindUpdate(t, `{"title": "new", "caption": null, "is_hidden": true}`).ToPatch()

	if !patch.Title.Set || patch.Title.Value == nil || *patch.Title.Value != "new" {
		t.Errorf("title = %+v, want set to \"new\"", patch.Title)
	}
	if !patch.Caption.Set || patch.Caption.Value != nil {
		t.Errorf("caption = %+v, want set to nil (clear)", patch.Caption)
	}
	if patch.TextData.Set || patch.OCRText.Set || patch.Link.Set {
		t.Errorf("omitted fields must stay unset: text_data=%+v ocr_text=%+v link=%+v", patch.TextData, patch.OCRText, patch.Link)
	}
	if patch.Type.Set {
		t.Errorf("type = %+v, want unset", patch.Type)
	}
	if !patch.IsHidden.Set || !patch.IsHidden.Value {
		t.Errorf("is_hidden = %+v, want set to true", patch.IsHidden)
	}
}

func TestUpdateContentRequestRejectsWrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/contents/1", strings.NewReader(`{"title": 42}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req UpdateContentRequest
	if err := c.ShouldBind(&req); err == nil {
		t.Fatal("ShouldBind error = nil, want error for non-string title")
	}
}
