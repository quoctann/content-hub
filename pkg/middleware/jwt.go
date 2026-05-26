package middleware

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quoctann/content-hub/pkg/server"
)

type JWTAuthConfig struct {
	Secret string
	Expiry time.Duration
	Issuer string
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
}

func JWTAuth(cfg JWTAuthConfig) server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			return nil, server.Error401("missing access token")
		}

		claims, err := ParseAndValidateToken(tokenString, cfg.Secret)
		if err != nil {
			return nil, err
		}

		if claims.TokenType != "access" {
			return nil, server.Error401("invalid token type")
		}

		if claims.Role != "admin" {
			return nil, server.Error403("forbidden")
		}

		userID, _ := strconv.ParseInt(claims.Subject, 10, 64)
		c.SetValue("userID", userID)
		c.SetValue("role", claims.Role)

		return c, nil
	}
}

func RefreshTokenAuth(secret string) server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		tokenString, err := c.Cookie("refresh_token")
		if err != nil || tokenString == "" {
			return nil, server.Error401("missing refresh token")
		}

		claims, err := ParseAndValidateToken(tokenString, secret)
		if err != nil {
			return nil, err
		}

		if claims.TokenType != "refresh" {
			return nil, server.Error401("invalid token type")
		}

		userID, _ := strconv.ParseInt(claims.Subject, 10, 64)
		c.SetValue("userID", userID)
		c.SetValue("role", claims.Role)

		return c, nil
	}
}

func ParseAndValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, server.Error401("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, server.Error401("invalid token claims")
	}

	return claims, nil
}

func GenerateToken(secret string, expiry time.Duration, issuer string, role string, tokenType string, userID int64) (string, error) {
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
		Role:      role,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
