package http

import (
	"__MODULE__/internal/entity/user"
)

type CreateUserRequestDTO struct {
	Username user.Username `json:"username" validate:"required"`
	Email    user.Email    `json:"email" validate:"required"`
	Phone    user.Phone    `json:"phone,omitempty"`
	Website  user.Website  `json:"website,omitempty"`
}

// GetUsersRequest is the HTTP request DTO for GET /users
// - page is optional and must be >= 1 when present
type GetUsersRequest struct {
	Page int `query:"page" validate:"omitempty,min=1"`
}

// UserResponse is the HTTP response DTO for a single user
type UserResponse struct {
	ID       user.ID           `json:"id,omitempty"`
	Name     user.FullName     `json:"name,omitempty"`
	Username user.Username     `json:"username,omitempty"`
	Email    user.Email        `json:"email,omitempty"`
	Phone    user.Phone        `json:"phone,omitempty"`
	Website  user.Website      `json:"website,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// UsersResponse is the top-level response
type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
	Meta  any            `json:"meta,omitempty"` // extend later if you want pagination data
}
