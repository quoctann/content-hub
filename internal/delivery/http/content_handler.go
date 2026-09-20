package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/server"
)

type ContentHandler struct {
	CUsecase domain.ContentUsecase
	Logger   logger.ILogger
}

func NewContentHandler(r server.Router, us domain.ContentUsecase, l logger.ILogger) {
	handler := &ContentHandler{
		CUsecase: us,
		Logger:   l,
	}
	r.POST("/contents", handler.Store)
	r.PUT("/contents/:id", handler.Update)
	r.GET("/contents", handler.Search)
}

func NewAdminContentHandler(r server.Router, us domain.ContentUsecase, l logger.ILogger) {
	handler := &ContentHandler{
		CUsecase: us,
		Logger:   l,
	}

	r.POST("", handler.Store)
	r.PUT("/:id", handler.Update)
	r.GET("", handler.AdminList)
	r.GET("/:id", handler.GetByID)
	r.DELETE("/:id", handler.Delete)
	r.DELETE("/bulk-delete", handler.BulkDelete)
	r.PATCH("/:id/hide", handler.ToggleHide)
}

// validContentTypes defines the allowed content type values for the type filter.
var validContentTypes = map[domain.ContentType]bool{
	domain.Text:  true,
	domain.Image: true,
	domain.Video: true,
}

const maxInt64 = int64(1<<63 - 1)

func invalidRequest(c server.Context, err error) {
	c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
}

func parseKeywords(kw string) []string {
	if kw == "" {
		return nil
	}

	parts := strings.Split(kw, ",")
	out := parts[:0] // same underlying array, but zero length, filtering-in-place
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func parseContentType(typeStr string) (domain.ContentType, string) {
	ct := domain.ContentType(typeStr)
	if !validContentTypes[ct] {
		return "", "invalid type: must be 'text', 'image', or 'video'"
	}
	return ct, ""
}

// clampPageSize ensures pageSize is between 1 and max (inclusive).
func paginationOffset(page, pageSize int64) (int64, bool) {
	if page < 1 || pageSize < 1 || page > maxInt64/pageSize+1 {
		return 0, false
	}
	return (page - 1) * pageSize, true
}

// Search godoc
// @Summary      Search contents
// @Description  Search contents by keywords and type
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        keywords   query     string  false  "Comma-separated keywords: hello,world"
// @Param        match_type query     string  false  "Match type: and (all must match) or or (any match), default or"
// @Param        type       query     string  false  "Content type filter (image, text, video)"
// @Param        num        query     int     false  "Number of results"
// @Param        cursor     query     string  false  "Cursor for pagination"
// @Success      200        {object}  ContentSearchResponseWrapper
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [get]
func (h *ContentHandler) Search(c server.Context) {
	var req SearchContentRequest
	if err := c.BindQuery(&req); err != nil {
		invalidRequest(c, err)
		return
	}
	num := req.Num
	if num == 0 {
		num = 10
	}

	filter := domain.SearchFilter{}

	if keywords := parseKeywords(req.Keywords); keywords != nil {
		filter.Keywords = keywords
		filter.MatchType = req.MatchType
		if filter.MatchType == "" {
			filter.MatchType = "or"
		}
	}

	if typeStr := req.Type; typeStr != "" {
		ct, errMsg := parseContentType(typeStr)
		if errMsg != "" {
			c.JSON(http.StatusBadRequest, map[string]string{"error": errMsg})
			return
		}
		filter.ContentType = ct
	}

	result, err := h.CUsecase.Search(c.Request().Context(), filter, req.Cursor, num)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to search contents", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ToContentSearchResponse(result))
}

// Store godoc
// @Summary      Store content
// @Description  Store a new content
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        content  body      CreateContentRequest  true  "Content"
// @Success      201      {object}  ContentResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [post]
func (h *ContentHandler) Store(c server.Context) {
	var req CreateContentRequest
	if err := c.Bind(&req); err != nil {
		invalidRequest(c, err)
		return
	}
	content := req.ToDomain()

	if err := h.CUsecase.Create(c.Request().Context(), content); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to store content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, ToContentResponse(content))
}

// Update godoc
// @Summary      Update content
// @Description  Update existing content (supports partial updates)
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Content ID"
// @Param        content  body      UpdateContentRequest  true  "Content (partial fields allowed)"
// @Success      200      {object}  ContentResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents/{id} [put]
func (h *ContentHandler) Update(c server.Context) {
	var path contentIDRequest
	if err := c.BindURI(&path); err != nil {
		invalidRequest(c, err)
		return
	}
	id := path.ID

	var req UpdateContentRequest
	if err := c.Bind(&req); err != nil {
		invalidRequest(c, err)
		return
	}

	// Fetch the current content to merge updates
	existing, err := h.CUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to fetch existing content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	// Merge: only update fields that are provided in the request
	if req.Title != nil {
		existing.Title = req.Title
	}
	if req.TextData != nil {
		existing.TextData = req.TextData
	}
	if req.OCRText != nil {
		existing.OCRText = req.OCRText
	}
	if req.Caption != nil {
		existing.Caption = req.Caption
	}
	if req.Link != nil {
		existing.Link = req.Link
	}
	if req.Type != nil {
		existing.Type = domain.ContentType(*req.Type)
	}
	if req.IsHidden != nil {
		existing.IsHidden = *req.IsHidden
	}

	if err := h.CUsecase.Update(c.Request().Context(), existing); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to update content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	updatedContent, err := h.CUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to fetch updated content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ToContentResponse(updatedContent))
}

// AdminList godoc
// @Summary      List all content for admin
// @Description  List all content including hidden ones, with optional search/filter
// @Tags         admin-contents
// @Accept      json
// @Produce    json
// @Param       page      query  int     false  "Page number (1-based)"
// @Param       page_size query  int     false  "Items per page (default 20, max 100)"
// @Param       keywords  query  string  false  "Comma-separated keywords: hello,world"
// @Param       type      query  string  false  "Content type filter (image, text, video)"
// @Param       visible   query  string  false  "Visibility filter: true (visible only), false (hidden only), or empty (all)"
// @Success    200      {object}  AdminContentResponseWrapper
// @Failure    500      {object}  map[string]string
// @Security   ApiKeyAuth
// @Router    /admin/contents [get]
func (h *ContentHandler) AdminList(ctx server.Context) {
	var req AdminListContentRequest
	if err := ctx.BindQuery(&req); err != nil {
		invalidRequest(ctx, err)
		return
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20 // default
	}

	filter := domain.SearchFilter{IncludeHidden: true}
	if keywords := parseKeywords(req.Keywords); keywords != nil {
		filter.Keywords = keywords
		filter.MatchType = req.MatchType
		if filter.MatchType == "" {
			filter.MatchType = "or"
		}
	}

	if typeStr := req.Type; typeStr != "" {
		ct, errMsg := parseContentType(typeStr)
		if errMsg != "" {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": errMsg})
			return
		}
		filter.ContentType = ct
	}

	if vis := req.Visible; vis != "" {
		filter.VisibilityFilter = vis
	}

	offset, ok := paginationOffset(page, pageSize)
	if !ok {
		invalidRequest(ctx, nil)
		return
	}
	cursor := strconv.FormatInt(offset, 10)

	result, err := h.CUsecase.Search(ctx.Request().Context(), filter, cursor, pageSize)
	if err != nil {
		h.Logger.Error(ctx.Request().Context(), "failed to list contents", logger.Error(err))
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, ToAdminContentResponse(result, page, pageSize))
}

// GetByID godoc
// @Summary      Get content by ID
// @Description  Get a single content entry by its ID
// @Tags         admin-contents
// @Accept       json
// @Produce      json
// @Param        id      path      int  true  "Content ID"
// @Success      200     {object}  ContentResponse
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /admin/contents/{id} [get]
func (h *ContentHandler) GetByID(c server.Context) {
	var path contentIDRequest
	if err := c.BindURI(&path); err != nil {
		invalidRequest(c, err)
		return
	}
	id := path.ID

	content, err := h.CUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to get content", logger.Error(err))
		c.JSON(http.StatusNotFound, map[string]string{"error": "content not found"})
		return
	}

	c.JSON(http.StatusOK, ToContentResponse(content))
}

// Delete godoc
// @Summary      Delete content
// @Description  Permanently delete a content entry
// @Tags         admin-contents
// @Accept       json
// @Produce      json
// @Param        id      path      int  true  "Content ID"
// @Success      204     "No Content"
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /admin/contents/{id} [delete]
func (h *ContentHandler) Delete(c server.Context) {
	var path contentIDRequest
	if err := c.BindURI(&path); err != nil {
		invalidRequest(c, err)
		return
	}
	id := path.ID

	if err := h.CUsecase.Delete(c.Request().Context(), id); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to delete content", logger.Error(err))
		c.JSON(http.StatusNotFound, map[string]string{"error": "content not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleHide godoc
// @Summary      Toggle content visibility
// @Description  Set content as hidden or visible
// @Tags         admin-contents
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "Content ID"
// @Param        request  body      ToggleHideRequest  true  "Hidden state"
// @Success      200      {object}  ContentResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /admin/contents/{id}/hide [patch]
func (h *ContentHandler) ToggleHide(c server.Context) {
	var path contentIDRequest
	if err := c.BindURI(&path); err != nil {
		invalidRequest(c, err)
		return
	}
	id := path.ID

	var req ToggleHideRequest
	if err := c.Bind(&req); err != nil {
		invalidRequest(c, err)
		return
	}

	if err := h.CUsecase.SetHidden(c.Request().Context(), id, *req.Hidden); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to toggle hide", logger.Error(err))
		c.JSON(http.StatusNotFound, map[string]string{"error": "content not found"})
		return
	}

	content, err := h.CUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to get content after toggle", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ToContentResponse(content))
}

func (h *ContentHandler) BulkDelete(c server.Context) {
	var req BulkDeleteRequest
	if err := c.Bind(&req); err != nil {
		invalidRequest(c, err)
		return
	}

	if err := h.CUsecase.DeleteMany(c.Request().Context(), req.IDs); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to bulk delete contents", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
