package mapper

import (
	adapter "__MODULE__/internal/dto/adapter/http"
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/dto/usecase"
	"__MODULE__/internal/entity/user"
	"strings"
)

func UserUsecaseToRepo(in usecase.BaseUser) repository.BaseUser {
	return repository.BaseUser{
		ID:       copyPtr(in.ID),
		FullName: copyPtr(in.FullName),
		Username: copyPtr(in.Username),
		Email:    copyPtr(in.Email),
		Avatar:   copyPtr(in.Avatar),
		Phone:    copyPtr(in.Phone),
		Website:  copyPtr(in.Website),

		Company:   nil,
		City:      nil,
		IsActive:  nil,
		CreatedAt: nil,
		UpdatedAt: nil,
	}
}

func UserRepoToUsecase(in repository.BaseUser) usecase.BaseUser {
	return usecase.BaseUser{
		ID:       copyPtr(in.ID),
		FullName: copyPtr(in.FullName),
		Username: copyPtr(in.Username),
		Email:    copyPtr(in.Email),
		Avatar:   copyPtr(in.Avatar),
		Phone:    copyPtr(in.Phone),
		Website:  copyPtr(in.Website),
	}
}

func UserIntegrationToUsecase(u integration.UserDTO) usecase.BaseUser {
	return usecase.BaseUser{
		ID:       ptrIfNotEmpty(u.ID),
		FullName: ptrIfNotEmpty(u.Name),
		Username: ptrIfNotEmpty(u.Username),
		Email:    ptrIfNotEmpty(u.Email),
		Phone:    ptrIfNotEmpty(u.Phone),
		Website:  ptrIfNotEmpty(u.Website),
	}
}

func CreateUserRequestDTOToBaseUser(req adapter.CreateUserRequestDTO) usecase.BaseUser {
	return usecase.BaseUser{
		Email:    ptr(user.Email(req.Email)),                // req.Email is openapi_types.Email (assume compatible with string)
		Username: ptr(user.Username(req.Username)),          // req.Username is string
		Phone:    ptrIfNotEmpty(user.Phone(*req.Phone)),     // req.Phone is *string
		Website:  ptrIfNotEmpty(user.Website(*req.Website)), // req.Website is *string
		// ID, FullName, Avatar remain nil
	}
}

// Updated UserUsecaseToIntegration to match adapter.UserResponse pointer fields
func UserUsecaseToIntegration(b usecase.BaseUser) adapter.UserResponse {
	return adapter.UserResponse{
		Id:       ptrIfNotEmpty(getString(b.ID)),
		Name:     ptrIfNotEmpty(getString(b.FullName)),
		Username: ptrIfNotEmpty(getString(b.Username)),
		Email:    ptrIfNotEmpty(getString(b.Email)),
		Phone:    ptrIfNotEmpty(getString(b.Phone)),
		Website:  ptrIfNotEmpty(getString(b.Website)),
		Extra:    &map[string]string{}, // extend when needed
	}
}

func copyPtr[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func ptr[T any](v T) *T { return &v }

func ptrIfNotEmpty[T ~string](v T) *T {
	if strings.TrimSpace(string(v)) == "" {
		return nil
	}
	return ptr(v)
}

// --- package-level generic helper (must NOT be a function literal) ---
func getString[T ~string](p *T) string {
	if p == nil {
		return ""
	}
	return string(*p)
}
