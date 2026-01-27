package http

import (
	"net/http"
	"strconv"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/server"
)

type ArticleHandler struct {
	AUsecase domain.ArticleUsecase
}

func NewArticleHandler(r server.Router, us domain.ArticleUsecase) {
	handler := &ArticleHandler{
		AUsecase: us,
	}
	r.POST("/articles", handler.Store)
	r.GET("/articles/:id", handler.GetByID)
	r.GET("/articles", handler.Fetch)
}

// Fetch godoc
// @Summary      Fetch articles
// @Description  Fetch articles with pagination
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        num    query     int     false  "Number of articles"
// @Param        cursor query     string  false  "Cursor for pagination"
// @Success      200    {array}   domain.Article
// @Failure      500    {object}  map[string]string
// @Router       /articles [get]
func (a *ArticleHandler) Fetch(c server.Context) {
	numS := c.Query("num")
	num, _ := strconv.ParseInt(numS, 10, 64)
	cursor := c.Query("cursor")

	listAr, err := a.AUsecase.Fetch(c.Request().Context(), cursor, num)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, listAr)
}

// GetByID godoc
// @Summary      Get article by ID
// @Description  Get article by ID
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Article ID"
// @Success      200  {object}  domain.Article
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /articles/{id} [get]
func (a *ArticleHandler) GetByID(c server.Context) {
	idS := c.Param("id")
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	art, err := a.AUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]string{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, art)
}

// Store godoc
// @Summary      Store article
// @Description  Store a new article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        article  body      domain.Article  true  "Article"
// @Success      201      {object}  domain.Article
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /articles [post]
func (a *ArticleHandler) Store(c server.Context) {
	var article domain.Article
	if err := c.Bind(&article); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := a.AUsecase.Store(c.Request().Context(), &article); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, article)
}
