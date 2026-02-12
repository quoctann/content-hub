package http

import (
	"net/http"
	"strconv"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/server"
)

type ContentHandler struct {
	CUsecase domain.ContentUsecase
}

func NewContentHandler(r server.Router, us domain.ContentUsecase) {
	handler := &ContentHandler{
		CUsecase: us,
	}
	r.POST("/contents", handler.Store)
	r.PUT("/contents/:id", handler.Update)
	r.GET("/contents", handler.Search)
}

// Search godoc
// @Summary      Search contents
// @Description  Search contents by query text
// @Tags         contents
// @Accept       json
// @Produce      json
// @Param        q      query     string  false  "Search query"
// @Param        num    query     int     false  "Number of results"
// @Param        cursor query     string  false  "Cursor for pagination"
// @Success      200    {array}   domain.Content
// @Failure      500    {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /contents [get]
func (h *ContentHandler) Search(c server.Context) {
	query := c.Query("q")
	numS := c.Query("num")
	num, _ := strconv.ParseInt(numS, 10, 64)
	cursor := c.Query("cursor")

	contents, err := h.CUsecase.Search(c.Request().Context(), query, cursor, num)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}
