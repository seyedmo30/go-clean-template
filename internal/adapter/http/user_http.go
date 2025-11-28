package http

import (
	"net/http"

	"__MODULE__/internal/dto/usecase"
	"__MODULE__/internal/interfaces"

	"github.com/labstack/echo/v4"
)

// GetUsersRequest is the HTTP request DTO for GET /users
// - page is optional and must be >= 1 when present
type GetUsersRequest struct {
	Page int `query:"page" validate:"omitempty,min=1"`
}

// UserResponse is the HTTP response DTO for a single user
type UserResponse struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Username string            `json:"username,omitempty"`
	Email    string            `json:"email,omitempty"`
	Phone    string            `json:"phone,omitempty"`
	Website  string            `json:"website,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// UsersResponse is the top-level response
type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Meta  any            `json:"meta,omitempty"` // extend later if you want pagination data
}

// UserHandler handles HTTP endpoints for users.
type UserHandler struct {
	usecase interfaces.UserUsecase
}

// NewUserHandler constructs a handler.
func NewUserHandler(uc interfaces.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(c echo.Context) error {
	// Bind and validate
	var req GetUsersRequest
	// Note: Bind will decode query params into req.Page
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(&req); err != nil {
		// return human-friendly validation message
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// default page if not provided
	page := req.Page
	if page == 0 {
		page = 1
	}

	// call usecase
	users, err := h.usecase.GetUsers(c.Request().Context(), page)
	if err != nil {
		return handleUsecaseError(c, err)
	}
	// map to HTTP DTOs
	resp := UsersResponse{Users: make([]UserResponse, 0, len(users))}
	for _, u := range users {
		resp.Users = append(resp.Users, mapUsecaseBaseUserToHTTP(u))
	}

	return c.JSON(http.StatusOK, resp)
}

// helper to map usecase.BaseUser -> UserResponse
func mapUsecaseBaseUserToHTTP(b usecase.BaseUser) UserResponse {
	var id, name, username, email, phone, website string

	if b.ID != nil {
		id = string(*b.ID)
	}
	if b.FullName != nil {
		name = string(*b.FullName)
	}
	if b.Username != nil {
		username = string(*b.Username)
	}
	if b.Email != nil {
		email = string(*b.Email)
	}
	if b.Phone != nil {
		phone = string(*b.Phone)
	}
	if b.Website != nil {
		website = string(*b.Website)
	}

	// extra is not present in usecase.BaseUser; if you want to include company/city
	// you can extend BaseUser or load them here. We'll return empty Extra for now.
	extra := map[string]string{}

	// if you have company/city as part of usecase.BaseUser in future:
	// if b.Company != nil { extra["company"] = string(*b.Company) }

	return UserResponse{
		ID:       id,
		Name:     name,
		Username: username,
		Email:    email,
		Phone:    phone,
		Website:  website,
		Extra:    extra,
	}
}
