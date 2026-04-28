package usecase

import (
	"context"
	"time"

	"github.com/quoctann/content-hub/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type accountUsecase struct {
	repo           domain.AccountRepository
	jwtSecret      string
	jwtExpiry      time.Duration
	contextTimeout time.Duration
}

func NewAccountUsecase(repo domain.AccountRepository, jwtSecret string, jwtExpiry time.Duration) domain.AccountUsecase {
	return &accountUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (u *accountUsecase) Login(ctx context.Context, username, password string) (*domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	acc, err := u.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if !acc.IsActive {
		return nil, ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password)); err != nil {
		return nil, ErrUnauthorized
	}

	return acc, nil
}

func (u *accountUsecase) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return u.repo.GetByID(ctx, id)
}
