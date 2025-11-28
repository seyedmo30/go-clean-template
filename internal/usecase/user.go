package usecase

import (
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/dto/usecase"
	"context"
	"fmt"
	"time"
)

// GetUsers implements the flow:
// 1) query DB
// 2) if DB has rows -> map to usecase.BaseUser and return
// 3) otherwise call client, persist each user, and return mapped usecase.BaseUser list
func (u *userUsecase) GetUsers(ctx context.Context, page int) (res []usecase.BaseUser, err error) {
	if page <= 0 {
		page = 1
	}

	listReq := repository.ListRepositoryRequestDTO[repository.BaseUser]{
		BasePaginationRequest: repository.BasePaginationRequest{
			Limit: u.limit,
			Page:  page,
		},
	}

	// 1) query DB
	dbRes, err := u.repo.GetUsersList(ctx, listReq)
	if err == nil && len(dbRes.List) > 0 {
		out := make([]usecase.BaseUser, 0, len(dbRes.List))
		for _, bu := range dbRes.List {
			out = append(out, mapRepoBaseUserToUsecase(bu))
		}
		return out, nil
	}

	// propagate unexpected repository error
	if err != nil && len(dbRes.List) == 0 {
		return
	}

	// 2) query client
	clientResp, err := u.client.GetUsers(ctx, page)
	if err != nil {
		return
	}

	// 3) persist results
	now := time.Now()
	for _, cu := range clientResp.Users {
		r := repository.CreateUserRepositoryRequestDTO{
			BaseUser: mapIntegrationUserToRepo(cu, now),
		}
		if err := u.repo.CreateUser(ctx, r); err != nil {
			return nil, fmt.Errorf("repository.CreateUser: %w", err)
		}
	}

	// 4) map client response to usecase.BaseUser list and return
	out := make([]usecase.BaseUser, 0, len(clientResp.Users))
	for _, cu := range clientResp.Users {
		out = append(out, mapIntegrationToUsecaseBaseUser(cu))
	}
	return out, nil
}
