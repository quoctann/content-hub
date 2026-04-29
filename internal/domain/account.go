package domain

import (
	"context"
	"time"
)

type Account struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type AccountRepository interface {
	Create(ctx context.Context, acc *Account) error
	GetByID(ctx context.Context, id int64) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	Update(ctx context.Context, acc *Account) error
	Delete(ctx context.Context, id int64) error
}

type AccountUsecase interface {
	Login(ctx context.Context, username, password string) (*Account, error)
	GetByID(ctx context.Context, id int64) (*Account, error)
}
