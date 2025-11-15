package repository

import (
	"__MODULE__/internal/entity/user"
	"time"
)

type BaseUser struct {
	ID       *user.ID       `gorm:"primaryKey;column:id;type:text"`
	FullName *user.FullName `gorm:"column:full_name;type:text"`
	Username *user.Username `gorm:"column:username;type:text;index"`
	Email    *user.Email    `gorm:"column:email;type:text;index"`
	Avatar   *user.Avatar   `gorm:"column:avatar;type:text"`
	Phone    *user.Phone    `gorm:"column:phone;type:text"`
	Website  *user.Website  `gorm:"column:website;type:text"`
	Company  *user.Company  `gorm:"column:company;type:text"`
	City     *user.City     `gorm:"column:city;type:text"`

	IsActive  *bool      `gorm:"column:is_active"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

type CreateUserRepositoryRequestDTO struct {
	BaseUser
}

type UpdateUserRepositoryRequestDTO struct {
	BaseUser
}
