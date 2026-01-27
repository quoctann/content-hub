package usecase

import (
	"context"

	"github.com/quoctann/content-hub/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{
		userRepo: repo,
	}
}

func (u *userUsecase) FetchActive(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.FetchActive(ctx)
}

func (u *userUsecase) GetByID(ctx context.Context, id int64) (domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}
