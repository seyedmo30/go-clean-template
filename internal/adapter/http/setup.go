package http

import (
	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"
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
