package mapper

import (
	adapter "__MODULE__/internal/dto/adapter/http"
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/dto/usecase"
	entity "__MODULE__/internal/entity/user"
	"strings"
)

func copyPtr[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

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

func ptr[T any](v T) *T { return &v }

func ptrIfNotEmpty[T ~string](v T) *T {
	if strings.TrimSpace(string(v)) == "" {
		return nil
	}
	return ptr(v)
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

// --- package-level generic helper (must NOT be a function literal) ---
func getString[T ~string](p *T) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

// UserUsecaseToIntegration uses the package-level generic helper
func UserUsecaseToIntegration(b usecase.BaseUser) adapter.UserResponse {
	return adapter.UserResponse{
		ID:       entity.ID(getString(b.ID)),
		Name:     entity.FullName(getString(b.FullName)),
		Username: entity.Username(getString(b.Username)),
		Email:    entity.Email(getString(b.Email)),
		Phone:    entity.Phone(getString(b.Phone)),
		Website:  entity.Website(getString(b.Website)),
		Extra:    map[string]string{}, // extend when needed
	}
}
