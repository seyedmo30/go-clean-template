package usecase

import (
	"__MODULE__/internal/entity/user"
)

type BaseUser struct {
	ID       *user.ID
	FullName *user.FullName
	Username *user.Username
	Email    *user.Email
	Avatar   *user.Avatar
	Phone    *user.Phone
	Website  *user.Website
}
