package http

import (
	"__MODULE__/internal/interfaces"

	client "github.com/seyedmo30/go-clean-template-client/api"

	"github.com/labstack/echo/v4"
)

// RegisterUserRoutes registers user-related routes on the given Echo instance.
func RegisterUserRoutes(e *echo.Echo, uc interfaces.UserUsecase) {
	e.GET("/openapi/openapi.json", func(c echo.Context) error {
		return c.Blob(200, "application/json", client.OpenAPISpec)
	})
	e.Static("/docs", "assets/swagger")

	SetupValidator(e) // ensure validator is set

	h := NewUserHandler(uc)
	g := e.Group("/users")
	// GET /users?page=1
	g.GET("", h.GetUsers)
	g.POST("", h.CreateUsers)
	// you can add other endpoints: POST /users, GET /users/:id, etc.

}
