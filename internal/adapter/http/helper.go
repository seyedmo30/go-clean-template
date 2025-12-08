package http

import (
	"__MODULE__/pkg"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// handleUsecaseError centralizes error handling/logging for usecase errors.
// - c: echo context
// - err: the error from usecase
func handleUsecaseError(c echo.Context, err error) error {
	// prepare base log fields

	var appErr *pkg.AppError
	if errors.As(err, &appErr) {
		logData := logrus.Fields{}
		logData["stack"] = appErr.AppendStackLog(2).StackStr()
		logData["description"] = appErr.DescriptionStr()
		logData["metadata"] = appErr.Meta()
		logData["internal_code"] = appErr.InternalCode()
		logData["external_code"] = appErr.ExternalCode()

		logrus.WithFields(logData).Log(appErr.Level(), appErr.Message())

		return c.JSON(appErr.ExternalCode(), map[string]string{
			"error": appErr.Message(),
			"code":  appErr.InternalCodeStr(),
		})
	}

	// fallback: unexpected
	// TODO : append panic log
	// logrus.WithFields(logData).WithError(err).Error(err)
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
