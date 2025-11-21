package repository

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/entity/user"
	"__MODULE__/pkg"

	"github.com/go-sql-driver/mysql"
)

// The repository receiver used in your project. Replace with your actual repo type.
// type serviceRepository struct{ /* ... */ }

const maxPhoneLen = 20

func (r *serviceRepository) CreateUser(ctx context.Context, params repository.CreateUserRepositoryRequestDTO) error {
	if params.Phone != nil {
		// convert named string type to builtin string for processing
		phoneStr := strings.TrimSpace(string(*params.Phone))

		// rune-safe truncate
		if utf8.RuneCountInString(phoneStr) > maxPhoneLen {
			runes := []rune(phoneStr)
			phoneStr = string(runes[:maxPhoneLen])
		}

		// convert back to the named Phone type and write back
		*params.Phone = user.Phone(phoneStr)
	}

	if err := db.WithContext(ctx).Table("users").Create(&params).Error; err != nil {
		return r.handleDBErrors(err)
	}
	return nil
}

// GetUsersList returns a paginated list of users.
func (r *serviceRepository) GetUsersList(ctx context.Context, params repository.ListRepositoryRequestDTO[repository.BaseUser]) (res repository.ListRepositoryResponseDTO[repository.BaseUser], err error) {
	var items []repository.BaseUser
	offset := (params.Page - 1) * params.Limit
	var total int64

	// total count
	db.Table("users").Count(&total)

	query := db.Table("users").
		Limit(params.Limit).
		Offset(offset).
		Find(&items)

	if err := query.Error; err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			ErrMsg := fmt.Sprintf("%s %d", mysqlErr.Message, mysqlErr.Number)
			err = pkg.ErrInternalServerError.AddStack().AddDescription(ErrMsg)
			return res, err
		}
		return res, r.handleDBErrors(err)
	}

	hasMore := (offset + len(items)) < int(total)

	res = repository.ListRepositoryResponseDTO[repository.BaseUser]{
		BasePaginationResponse: repository.BasePaginationResponse{
			Limit:   params.Limit,
			Page:    params.Page,
			HasMore: hasMore,
			Total:   total,
		},
		List: items,
	}

	return res, nil
}

// GetUserById retrieves a single user by id.
func (r *serviceRepository) GetUserById(ctx context.Context, id string) (repository.BaseUser, error) {
	var user repository.BaseUser
	if err := db.Table("users").Where("id = ?", id).First(&user).Error; err != nil {
		return user, r.handleDBErrors(err)
	}
	return user, nil
}

// UpdateUser updates fields of an existing user.
func (r *serviceRepository) UpdateUser(ctx context.Context, params repository.UpdateUserRepositoryRequestDTO) error {
	// ensure IsActive has a value if nil (optional business rule)
	if params.IsActive == nil {
		params.IsActive = pkg.PtrBool(false)
	}

	result := db.Table("users").Where("id = ?", params.ID).Updates(&params)
	if err := result.Error; err != nil {
		return r.handleDBErrors(err)
	}
	if result.RowsAffected == 0 {
		return pkg.ErrRecordNotFound.AddStack().AddDescription("user not found")
	}
	return nil
}

// DeleteUser deletes a user by id.
func (r *serviceRepository) DeleteUser(ctx context.Context, id string) error {
	result := db.Table("users").Where("id = ?", id).Delete(&repository.BaseUser{})
	if err := result.Error; err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			ErrMsg := fmt.Sprintf("%s %d", mysqlErr.Message, mysqlErr.Number)
			err = pkg.ErrInternalServerError.AddStack().AddDescription(ErrMsg)
			return err
		}
		return r.handleDBErrors(err)
	}
	if result.RowsAffected == 0 {
		return pkg.ErrRecordNotFound.AddStack().AddDescription("user not found")
	}
	return nil
}
