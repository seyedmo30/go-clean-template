package integration

import (
	"__MODULE__/internal/entity/user"
)

// Unified DTOs used by the application.

type UserDTO struct {
	ID       user.ID           `json:"id"`
	Name     user.FullName     `json:"name,omitempty"`
	Username user.Username     `json:"username,omitempty"`
	Email    user.Email        `json:"email,omitempty"`
	Phone    user.Phone        `json:"phone,omitempty"`
	Website  user.Website      `json:"website,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"` // provider-specific fields
}

type MetaInfoDTO struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type UserListResponseDTO struct {
	Provider string      `json:"provider"`
	Users    []UserDTO   `json:"users"`
	Meta     MetaInfoDTO `json:"meta,omitempty"`
	Raw      []byte      `json:"-"`
}
