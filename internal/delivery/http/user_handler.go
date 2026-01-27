package http

import (
	"net/http"
	"strconv"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/server"
)

type UserHandler struct {
	UUsecase domain.UserUsecase
}

func NewUserHandler(r server.Router, us domain.UserUsecase) {
	handler := &UserHandler{
		UUsecase: us,
	}
	r.GET("/users", handler.FetchActive)
	r.GET("/users/:id", handler.GetByID)
}

// FetchActive godoc
// @Summary      Fetch active users
// @Description  Fetch all users with status 'active'
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200    {array}   domain.User
// @Failure      500    {object}  map[string]string
// @Router       /users [get]
func (u *UserHandler) FetchActive(c server.Context) {
	users, err := u.UUsecase.FetchActive(c.Request().Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetByID godoc
// @Summary      Get user by ID
// @Description  Get user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [get]
func (u *UserHandler) GetByID(c server.Context) {
	idS := c.Param("id")
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	user, err := u.UUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
