package middleware

import (
	"strings"
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
	Role string `json:"role"`
}

func JWTAuth(cfg JWTAuthConfig) server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return nil, server.Error401("missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, server.Error401("invalid authorization header format")
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})
		if err != nil || !token.Valid {
			return nil, server.Error401("invalid token")
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return nil, server.Error401("invalid token claims")
		}

		if claims.Role != "admin" {
			return nil, server.Error403("forbidden")
		}

		return c, nil
	}
}

func GenerateToken(secret string, expiry time.Duration, issuer string, role string) (string, error) {
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
