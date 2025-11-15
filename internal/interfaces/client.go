package interfaces

import (
	"context"

	"__MODULE__/internal/dto/client/integration"
)

// UserService represents an external user provider.
type UserService interface {
	// GetUsers fetches users. For providers with pagination, pass page>0.
	GetUsers(ctx context.Context, page int) (integration.UserListResponseDTO, error)
}
