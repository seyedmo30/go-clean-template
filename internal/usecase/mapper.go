package usecase

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/usecase"
	"__MODULE__/internal/entity/user"
)

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
