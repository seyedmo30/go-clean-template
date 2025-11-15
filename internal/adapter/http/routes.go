package http

import (
	"__MODULE__/internal/interfaces"

	"github.com/labstack/echo/v4"
)

// RegisterUserRoutes registers user-related routes on the given Echo instance.
func RegisterUserRoutes(e *echo.Echo, uc interfaces.UserUsecase) {
	SetupValidator(e) // ensure validator is set

	h := NewUserHandler(uc)
	g := e.Group("/users")
	// GET /users?page=1
	g.GET("", h.GetUsers)
	// you can add other endpoints: POST /users, GET /users/:id, etc.
}
