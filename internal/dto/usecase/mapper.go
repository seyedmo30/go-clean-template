package usecase

import (
	adapter "__MODULE__/internal/dto/adapter/http"
	"__MODULE__/internal/entity/user"
	"strings"
)

func ptr[T any](v T) *T { return &v }

// ptrIfNotEmpty returns a pointer to the string-like value if it's non-empty
// otherwise returns nil. We use a type constraint so it works with aliases
// whose underlying type is string (e.g. user.Username).
func ptrIfNotEmpty[T ~string](v T) *T {
	if strings.TrimSpace(string(v)) == "" {
		return nil
	}
	return ptr(v)
}

// MapCreateUser maps an adapter-level CreateUserRequestDTO into the
// usecase-level CreateUserRequestDTO. This keeps controller/adapter logic
// (HTTP concerns) separated from business/usecase boundary.
//
// Best practices shown here:
// - Perform minimal transformation (no business logic) — only shape conversion
// - Keep optional fields nil when missing (so usecase can decide defaulting)
// - Preserve domain types (use entity/user types in usecase DTOs)
// - Use small, well-named helper functions to keep mapping readable

func MapCreateUser(req adapter.CreateUserRequestDTO) CreateUserRequestDTO {
	return CreateUserRequestDTO{
		BaseUser: BaseUser{
			// Required fields: use ptrIfNotEmpty to keep them non-nil only when present.
			Username: ptrIfNotEmpty(req.Username),
			Email:    ptrIfNotEmpty(req.Email),

			// Optional fields: set only when non-empty so they remain nil otherwise.
			Phone: func() *user.Phone {
				if strings.TrimSpace(string(req.Phone)) == "" {
					return nil
				}
				return ptr(req.Phone)
			}(),

			Website: ptrIfNotEmpty(req.Website),
		},
	}
}
