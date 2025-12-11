package http

import (
	"net/http"

	adapter "__MODULE__/internal/dto/adapter/http"
	"__MODULE__/internal/dto/mapper"
	"__MODULE__/internal/dto/usecase"
	"__MODULE__/internal/interfaces"

	"github.com/labstack/echo/v4"
)

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
	var params adapter.GetUsersParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	page := 1
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}

	users, err := h.usecase.GetUsers(c.Request().Context(), page)
	if err != nil {
		return handleUsecaseError(c, err)
	}

	usersResp := make([]adapter.UserResponse, 0, len(users))
	for _, u := range users {
		usersResp = append(usersResp, mapper.UserUsecaseToIntegration(u))
	}

	resp := adapter.GetUsersResponse{Users: &usersResp}
	return c.JSON(http.StatusOK, resp)
}

// CreateUsers handles POST /users
func (h *UserHandler) CreateUsers(c echo.Context) error {
	var createReq adapter.CreateUserRequestDTO
	if err := c.Bind(&createReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(&createReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	baseUser := mapper.CreateUserRequestDTOToBaseUser(createReq)
	ucReq := usecase.CreateUserRequestDTO{BaseUser: baseUser}

	createdUsers, err := h.usecase.CreateUser(c.Request().Context(), ucReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	createdUsersResp := make([]adapter.UserResponse, 0, len(createdUsers))
	for _, u := range createdUsers {
		createdUsersResp = append(createdUsersResp, mapper.UserUsecaseToIntegration(u))
	}

	return c.JSON(http.StatusOK, createdUsersResp)
}
