package interfaces

import (
	"__MODULE__/internal/dto/repository"
	"context"
)

// Repository is an interface that defines the methods for interacting with the repository.
type Repository interface {
	CreateUser(ctx context.Context, params repository.CreateUserRepositoryRequestDTO) error
	GetUsersList(ctx context.Context, params repository.ListRepositoryRequestDTO[repository.BaseUser]) (res repository.ListRepositoryResponseDTO[repository.BaseUser], err error)
	GetUserById(ctx context.Context, id string) (repository.BaseUser, error)
	UpdateUser(ctx context.Context, params repository.UpdateUserRepositoryRequestDTO) error
	DeleteUser(ctx context.Context, id string) error
}
