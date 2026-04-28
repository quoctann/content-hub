package http

import (
	"net/http"
	"time"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/server"
)

type AccountHandler struct {
	AUsecase domain.AccountUsecase
	Config   *config.Config
	Logger   logger.ILogger
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewAccountHandler(r server.Router, au domain.AccountUsecase, cfg *config.Config, l logger.ILogger) {
	handler := &AccountHandler{
		AUsecase: au,
		Config:   cfg,
		Logger:   l,
	}
	r.POST("/account/login", handler.Login)
}

func (h *AccountHandler) Login(c server.Context) {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	account, err := h.AUsecase.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		h.Logger.Warn(c.Request().Context(), "failed login attempt", logger.String("username", req.Username), logger.Error(err))
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	expiry, err := time.ParseDuration(h.Config.Security.JWTExpiry)
	if err != nil {
		expiry = 24 * time.Hour
	}

	token, err := middleware.GenerateToken(
		h.Config.Security.JWTSecret,
		expiry,
		"content-hub",
		account.Role,
	)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	refreshToken, err := middleware.GenerateToken(
		h.Config.Security.JWTSecret,
		expiry*7,
		"content-hub",
		account.Role,
	)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate refresh token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiry.Seconds()),
	})
}
