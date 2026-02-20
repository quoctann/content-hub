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

// validContentTypes defines the allowed content type values for the type filter.
var validContentTypes = map[domain.ContentType]bool{
	domain.Text:  true,
	domain.Image: true,
}

// Search godoc
// @Summary      Search contents
// @Description  Search contents by keywords, query text, type, and tags
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        keywords   query     string  false  "Comma-separated keywords: hello,world"
// @Param        match_type query     string  false  "Match type: and (all must match) or or (any match), default or"
// @Param        q          query     string  false  "Search query (deprecated, use keywords)"
// @Param        type       query     string  false  "Content type filter (image, text)"
// @Param        num        query     int     false  "Number of results"
// @Param        cursor     query     string  false  "Cursor for pagination"
// @Success      200        {object}  ContentSearchResponseWrapper
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [get]
func (h *ContentHandler) Search(c server.Context) {
	numS := c.Query("num")
	num, _ := strconv.ParseInt(numS, 10, 64)
	cursor := c.Query("cursor")

	// Build search filter
	filter := domain.SearchFilter{}

	// Parse keywords (preferred) or fallback to legacy q
	keywordsStr := c.Query("keywords")
	if keywordsStr != "" {
		parts := strings.Split(keywordsStr, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				filter.Keywords = append(filter.Keywords, trimmed)
			}
		}
		// Default to "or" if match_type not provided
		filter.MatchType = c.Query("match_type")
		if filter.MatchType == "" {
			filter.MatchType = "or"
		}
	} else {
		// Legacy: fallback to q query
		filter.Query = c.Query("q")
	}

	// Validate and set content type filter
	if typeStr := c.Query("type"); typeStr != "" {
		ct := domain.ContentType(typeStr)
		if !validContentTypes[ct] {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid type: must be 'image' or 'text'"})
			return
		}
		filter.ContentType = ct
	}

	result, err := h.CUsecase.Search(c.Request().Context(), filter, cursor, num)
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
// @Param        content  body      domain.Content  true  "Content"
// @Success      201      {object}  ContentResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [post]
func (h *ContentHandler) Store(c server.Context) {
	var content domain.Content
	if err := c.Bind(&content); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := h.CUsecase.Create(c.Request().Context(), &content); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to store content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, ToContentResponse(&content))
}

// Update godoc
// @Summary      Update content
// @Description  Update existing content
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Content ID"
// @Param        content  body      domain.Content  true  "Content"
// @Success      200      {object}  ContentResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents/{id} [put]
func (h *ContentHandler) Update(c server.Context) {
	idS := c.Param("id")
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	var content domain.Content
	if err := c.Bind(&content); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	content.ID = id

	if err := h.CUsecase.Update(c.Request().Context(), &content); err != nil {
		h.Logger.Error(c.Request().Context(), "failed to update content", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ToContentResponse(&content))
}
