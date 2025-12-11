package http

import (
	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"

	api "__MODULE__/internal/dto/adapter/http"
	"encoding/json"
	"net/http"
)

// EchoValidator is a thin wrapper to integrate go-playground/validator with Echo.
type EchoValidator struct {
	validator *validator.Validate
}

func NewEchoValidator() *EchoValidator {
	return &EchoValidator{validator: validator.New()}
}

func (v *EchoValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// SetupValidator attaches validator to echo instance. Call once at server startup.
func SetupValidator(e *echo.Echo) {
	e.Validator = NewEchoValidator()
}

func serveSpec(c echo.Context) error {
	swagger, err := api.GetSwagger()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to load swagger")
	}
	// Marshal back to JSON and return (or use as needed)
	bs, _ := json.Marshal(swagger)
	return c.Blob(http.StatusOK, "application/json", bs)
}
