package interfaces

import (
	"__MODULE__/internal/dto/usecase"
	"context"
)

// UserUsecase defines available usecase methods for users.
type UserUsecase interface {
	CreateUser(ctx context.Context, req usecase.CreateUserRequestDTO) ([]usecase.BaseUser, error)
	// GetUsers returns users for given page.
	// First tries repository; if nothing found calls external client, persists results and returns them.
	GetUsers(ctx context.Context, page int) ([]usecase.BaseUser, error)
}

type BackgroundJobUsecase interface{}
