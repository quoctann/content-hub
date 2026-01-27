package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserRepository interface {
	FetchActive(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id int64) (User, error)
}

type UserUsecase interface {
	FetchActive(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id int64) (User, error)
}
