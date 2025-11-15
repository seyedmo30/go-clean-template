package interfaces

import (
	"__MODULE__/internal/dto/client/integration"
	"context"
)

// UserUsecase defines available usecase methods for users.
type UserUsecase interface {
	// GetUsers returns users for given page.
	// First tries repository; if nothing found calls external client, persists results and returns them.
	GetUsers(ctx context.Context, page int) (integration.UserListResponseDTO, error)
}
