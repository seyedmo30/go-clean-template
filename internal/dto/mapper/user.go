package mapper

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/dto/usecase"
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
