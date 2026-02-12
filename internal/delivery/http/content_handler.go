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
// @Description  Search contents by query text, type, and tags
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        q      query     string  false  "Search query"
// @Param        type   query     string  false  "Content type filter (image, text)"
// @Param        tags   query     string  false  "Comma-separated tag names"
// @Param        num    query     int     false  "Number of results"
// @Param        cursor query     string  false  "Cursor for pagination"
// @Success      200    {array}   domain.Content
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [get]
func (h *ContentHandler) Search(c server.Context) {
	numS := c.Query("num")
	num, _ := strconv.ParseInt(numS, 10, 64)
	cursor := c.Query("cursor")

	// Build search filter
	filter := domain.SearchFilter{
		Query: c.Query("q"),
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

	// Parse comma-separated tags
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		filter.Tags = tags
	}

	contents, err := h.CUsecase.Search(c.Request().Context(), filter, cursor, num)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to search contents", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, contents)
}

// Store godoc
// @Summary      Store content
// @Description  Store a new content
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        content  body      domain.Content  true  "Content"
// @Success      201      {object}  domain.Content
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

	c.JSON(http.StatusCreated, content)
}

// Update godoc
// @Summary      Update content
// @Description  Update existing content
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Content ID"
// @Param        content  body      domain.Content  true  "Content"
// @Success      200      {object}  domain.Content
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

	c.JSON(http.StatusOK, content)
}
