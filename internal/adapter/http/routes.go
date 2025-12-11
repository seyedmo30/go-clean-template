package http

import (
	api "__MODULE__/internal/dto/adapter/http"
	"__MODULE__/internal/interfaces"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RegisterUserRoutes registers user-related routes on the given Echo instance.
func RegisterUserRoutes(e *echo.Echo, uc interfaces.UserUsecase) {
	SetupValidator(e) // ensure validator is set

	h := NewUserHandler(uc)
	g := e.Group("/users")
	// GET /users?page=1
	g.GET("", h.GetUsers)
	g.POST("", h.CreateUsers)
	// you can add other endpoints: POST /users, GET /users/:id, etc.
}

// RegisterSwagger registers routes to serve the OpenAPI spec and a quick docs redirect
func RegisterSwagger(e *echo.Echo) {
	e.GET("/openapi.json", func(c echo.Context) error {
		swagger, err := api.GetSwagger()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to load swagger")
		}
		bs, err := json.Marshal(swagger)
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to marshal swagger")
		}
		return c.Blob(http.StatusOK, "application/json", bs)
	})

	// quick redirect to Petstore UI (pointing to your spec)
	// (Useful in development; you can host a local swagger-ui if you prefer)
	e.GET("/docs", func(c echo.Context) error {
		// Update host+port if not using 8080
		return c.Redirect(http.StatusFound, "https://petstore.swagger.io/?url=http://localhost:8009/openapi.json")
	})
}
