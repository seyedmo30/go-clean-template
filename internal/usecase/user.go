package usecase

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/entity/user"
	"context"
	"fmt"
	"time"
)

// GetUsers implements the described flow:
// 1. Try repository.GetUsersList
// 2. If repository returns items -> map and return
// 3. Otherwise call client.GetUsers, persist each item via repository.CreateUser, and return client response
func (u *userUsecase) GetUsers(ctx context.Context, page int) (integration.UserListResponseDTO, error) {
	// prepare repository list request
	if page <= 0 {
		page = 1
	}
	listReq := repository.ListRepositoryRequestDTO[repository.BaseUser]{
		BasePaginationRequest: repository.BasePaginationRequest{
			Limit: u.limit,
			Page:  page,
		},
		// Filter left zero-valued for "no filter" (change if you want to filter)
	}

	// 1) query DB
	dbRes, err := u.repo.GetUsersList(ctx, listReq)
	if err == nil && len(dbRes.List) > 0 {
		// map repository.BaseUser -> integration.UserDTO
		users := make([]integration.UserDTO, 0, len(dbRes.List))
		for _, bu := range dbRes.List {
			users = append(users, mapRepoUserToIntegration(bu))
		}

		return integration.UserListResponseDTO{
			Provider: "db",
			Users:    users,
			Meta: integration.MetaInfoDTO{
				Page:    page,
				PerPage: dbRes.Limit,
				Total:   int(dbRes.Total),
			},
		}, nil
	}

	// if repository returned an error other than "not found", propagate it
	// Note: adjust as needed if your repo returns a specific not-found error type.
	if err != nil && len(dbRes.List) == 0 {
		// If repository error is expected to be tolerated, you can log and continue.
		// Here we return error to avoid masking unexpected DB failure.
		return integration.UserListResponseDTO{}, fmt.Errorf("repository.GetUsersList: %w", err)
	}

	// 2) repository empty -> call external client
	clientResp, err := u.client.GetUsers(ctx, page)
	if err != nil {
		return integration.UserListResponseDTO{}, fmt.Errorf("client.GetUsers: %w", err)
	}

	// 3) persist client users to repo
	now := time.Now()
	for _, cu := range clientResp.Users {
		r := repository.CreateUserRepositoryRequestDTO{
			BaseUser: mapIntegrationUserToRepo(cu, now),
		}
		if err := u.repo.CreateUser(ctx, r); err != nil {
			// fail fast - return error so caller knows persistence failed
			return integration.UserListResponseDTO{}, fmt.Errorf("repository.CreateUser: %w", err)
		}
	}

	// Return the client response as-is (already in integration.UserListResponseDTO)
	return clientResp, nil
}

// ---------- helpers: mapping between repository.BaseUser <-> integration.UserDTO ----------

func mapRepoUserToIntegration(b repository.BaseUser) integration.UserDTO {
	var id user.ID
	var name user.FullName
	var uname user.Username
	var email user.Email
	var phone user.Phone
	var website user.Website
	var extra map[string]string

	if b.ID != nil {
		id = *b.ID
	}
	if b.FullName != nil {
		name = *b.FullName
	}
	if b.Username != nil {
		uname = *b.Username
	}
	if b.Email != nil {
		email = *b.Email
	}
	if b.Phone != nil {
		phone = *b.Phone
	}
	if b.Website != nil {
		website = *b.Website
	}
	extra = make(map[string]string)
	if b.Avatar != nil {
		extra["avatar"] = string(*b.Avatar)
	}
	if b.Company != nil {
		extra["company"] = string(*b.Company)
	}
	if b.City != nil {
		extra["city"] = string(*b.City)
	}

	return integration.UserDTO{
		ID:       id,
		Name:     name,
		Username: uname,
		Email:    email,
		Phone:    phone,
		Website:  website,
		Extra:    extra,
	}
}

func mapIntegrationUserToRepo(u integration.UserDTO, now time.Time) repository.BaseUser {
	// create local copies for pointer assignment
	var (
		id       = u.ID
		name     = u.Name
		username = u.Username
		email    = u.Email
		phone    = u.Phone
		website  = u.Website
		isActive = true
		created  = now
		updated  = now
	)

	var avatar user.Avatar
	var company user.Company
	var city user.City

	// optional extra fields
	if v, ok := u.Extra["avatar"]; ok {
		avatar = user.Avatar(v)
	}
	if v, ok := u.Extra["company"]; ok {
		company = user.Company(v)
	}
	if v, ok := u.Extra["city"]; ok {
		city = user.City(v)
	}

	// set nil pointers for zero-values to keep DB columns null when absent
	var avatarPtr *user.Avatar
	var companyPtr *user.Company
	var cityPtr *user.City

	if avatar != "" {
		avatarPtr = &avatar
	}
	if company != "" {
		companyPtr = &company
	}
	if city != "" {
		cityPtr = &city
	}

	return repository.BaseUser{
		ID:        &id,
		FullName:  &name,
		Username:  &username,
		Email:     &email,
		Avatar:    avatarPtr,
		Phone:     &phone,
		Website:   &website,
		Company:   companyPtr,
		City:      cityPtr,
		IsActive:  &isActive,
		CreatedAt: &created,
		UpdatedAt: &updated,
	}
}
