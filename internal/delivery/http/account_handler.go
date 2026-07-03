package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

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
	CSRFToken string `json:"csrf_token"`
	ExpiresIn int64  `json:"expires_in"`
}

type RefreshResponse struct {
	CSRFToken string `json:"csrf_token"`
	ExpiresIn int64  `json:"expires_in"`
}

func NewAccountHandler(r server.RouterGroup, au domain.AccountUsecase, cfg *config.Config, l logger.ILogger) {
	handler := &AccountHandler{
		AUsecase: au,
		Config:   cfg,
		Logger:   l,
	}
	r.POST("/account/login", handler.Login)
	r.POST("/account/refresh", handler.Refresh)
	r.POST("/account/logout", handler.Logout)
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with username and password, returns access/refresh tokens in cookies
// @Tags         account
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest    true  "Login credentials"
// @Success      200      {object}  LoginResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /account/login [post]
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

	token, err := middleware.GenerateToken(
		h.Config.Security.JWTSecret,
		h.Config.Security.JWTExpiry,
		"content-hub",
		account.Role,
		"access",
		account.ID,
	)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	expiry := h.Config.Security.JWTExpiry
	refreshToken, err := middleware.GenerateToken(
		h.Config.Security.JWTSecret,
		expiry*7,
		"content-hub",
		account.Role,
		"refresh",
		account.ID,
	)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate refresh token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate csrf token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	isSecure := h.Config.Server.AppEnv != "local"

	c.SetCookie("access_token", token, int(expiry.Seconds()), "/", isSecure, true)
	c.SetCookie("refresh_token", refreshToken, int(expiry.Seconds()*7), "/", isSecure, true)
	c.SetCookie("csrf_token", csrfToken, int(expiry.Seconds()), "/", isSecure, false)

	c.JSON(http.StatusOK, LoginResponse{
		CSRFToken: csrfToken,
		ExpiresIn: int64(expiry.Seconds()),
	})
}

// Refresh godoc
// @Summary      Refresh token
// @Description  Refresh access token using refresh token from cookie
// @Tags         account
// @Accept       json
// @Produce      json
// @Success      200      {object}  RefreshResponse
// @Failure      401      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /account/refresh [post]
func (h *AccountHandler) Refresh(c server.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	claims, err := middleware.ParseAndValidateToken(refreshToken, h.Config.Security.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token type"})
		return
	}

	userID, _ := strconv.ParseInt(claims.Subject, 10, 64)

	token, err := middleware.GenerateToken(
		h.Config.Security.JWTSecret,
		h.Config.Security.JWTExpiry,
		"content-hub",
		claims.Role,
		"access",
		userID,
	)
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.Logger.Error(c.Request().Context(), "failed to generate csrf token", logger.Error(err))
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	isSecure := h.Config.Server.AppEnv != "local"

	expiry := h.Config.Security.JWTExpiry
	c.SetCookie("access_token", token, int(expiry.Seconds()), "/", isSecure, true)
	c.SetCookie("csrf_token", csrfToken, int(expiry.Seconds()), "/", isSecure, false)

	c.JSON(http.StatusOK, RefreshResponse{
		CSRFToken: csrfToken,
		ExpiresIn: int64(expiry.Seconds()),
	})
}

// Logout godoc
// @Summary      Logout
// @Description  Clear access/refresh tokens cookies
// @Tags         account
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /account/logout [post]
func (h *AccountHandler) Logout(c server.Context) {
	isSecure := h.Config.Server.AppEnv != "local"

	c.SetCookie("access_token", "", -1, "/", isSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", isSecure, true)
	c.SetCookie("csrf_token", "", -1, "/", isSecure, false)

	c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
