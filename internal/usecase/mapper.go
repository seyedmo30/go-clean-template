package usecase

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/dto/usecase"
	"__MODULE__/internal/entity/user"
	"time"
)

// ---------- mapping helpers: repository.BaseUser / integration.UserDTO -> usecase.BaseUser ----------

func mapRepoBaseUserToUsecase(b repository.BaseUser) usecase.BaseUser {
	// repository.BaseUser already uses pointer fields that match usecase.BaseUser, so copy them
	return usecase.BaseUser{
		ID:       b.ID,
		FullName: b.FullName,
		Username: b.Username,
		Email:    b.Email,
		Avatar:   b.Avatar,
		Phone:    b.Phone,
		Website:  b.Website,
	}
}
func mapIntegrationToUsecaseBaseUser(i integration.UserDTO) usecase.BaseUser {
	// optional avatar
	var avatarPtr *user.Avatar
	if v, ok := i.Extra["avatar"]; ok && v != "" {
		av := user.Avatar(v)
		avatarPtr = &av
	}

	return usecase.BaseUser{
		ID:       &i.ID,
		FullName: &i.Name,
		Username: &i.Username,
		Email:    &i.Email,
		Avatar:   avatarPtr,
		Phone:    &i.Phone,
		Website:  &i.Website,
	}
}

func mapIntegrationUserToRepo(u integration.UserDTO, now time.Time) repository.BaseUser {
	isActive := true

	// Optional fields
	var avatarPtr *user.Avatar
	var companyPtr *user.Company
	var cityPtr *user.City

	if v, ok := u.Extra["avatar"]; ok && v != "" {
		av := user.Avatar(v)
		avatarPtr = &av
	}
	if v, ok := u.Extra["company"]; ok && v != "" {
		cp := user.Company(v)
		companyPtr = &cp
	}
	if v, ok := u.Extra["city"]; ok && v != "" {
		ct := user.City(v)
		cityPtr = &ct
	}

	return repository.BaseUser{
		ID:       &u.ID,       // safe
		FullName: &u.Name,     // safe
		Username: &u.Username, // safe
		Email:    &u.Email,    // safe
		Phone:    &u.Phone,    // safe
		Website:  &u.Website,  // safe

		Avatar:  avatarPtr,
		Company: companyPtr,
		City:    cityPtr,

		IsActive:  &isActive,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}
