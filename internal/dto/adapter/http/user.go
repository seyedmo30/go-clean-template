package http

import "__MODULE__/internal/entity/user"



type CreateUserRequestDTO struct {
	Username user.Username `json:"username" validate:"required"`
	Email    user.Email    `json:"email" validate:"required"`
	Phone    user.Phone    `json:"phone,omitempty"`
	Website  user.Website  `json:"website,omitempty"`
}
