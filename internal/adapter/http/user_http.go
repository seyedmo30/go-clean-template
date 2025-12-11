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
	// Bind and validate
	var req adapter.GetUsersRequest
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
	resp := adapter.GetUsersResponse{Users: make([]adapter.UserResponse, 0, len(users))}
	for _, u := range users {
		resp.Users = append(resp.Users, mapper.UserUsecaseToIntegration(u))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) CreateUsers(c echo.Context) error {
	var req adapter.CreateUserRequestDTO

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(&req); err != nil {

		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	ucReq := usecase.MapCreateUser(req)
	users, err := h.usecase.CreateUser(c.Request().Context(), ucReq)

	if err != nil {

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	resp := adapter.GetUsersResponse{Users: make([]adapter.UserResponse, 0, len(users))}
	for _, u := range users {
		resp.Users = append(resp.Users, mapper.UserUsecaseToIntegration(u))
	}

	return c.JSON(http.StatusOK, resp)
}
