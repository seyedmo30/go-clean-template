package usecase

import (
	"__MODULE__/internal/interfaces"
)

type userUsecase struct {
	repo   interfaces.Repository  // persistence
	client interfaces.UserService // external user provider client
	limit  int                    // default page size
}

// NewUserUsecase creates a new instance of user usecase.
func NewUserUsecase(repo interfaces.Repository, client interfaces.UserService, defaultLimit int) userUsecase {
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	return userUsecase{
		repo:   repo,
		client: client,
		limit:  defaultLimit,
	}
}

var _ interfaces.UserUsecase = (*userUsecase)(nil)
