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

// ProviderService is the interface that defines the methods for managing bank providers.
// It allows registering new banks, getting a bank service by ID, stopping a bank service,
// and getting a list of supported providers.
type ProviderService interface {
}
