package pkg

import (
	"bytes"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

func init() {
	// Load the error configuration from the YAML file
	err := LoadErrorConfig("errors.json")
	if err != nil {
		panic("Failed to load error configuration: " + err.Error())
	}
}

var errorConfig map[string]AppError

var (
	ErrBadRequest *AppError
)

func LoadErrorConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &errorConfig)
	if err != nil {
		return err
	}
	initializeErrors()
	return nil
}

func initializeErrors() {
	ErrBadRequest = getErrorFromConfig("ErrBadRequest")

}

type AppError struct {
	Message        string `json:"message,omitempty"`
	LogDescription []byte `json:"description,omitempty"`
	LogStack       []byte `json:"stack,omitempty"`
	InternalCode   int
	ExternalCode   int
	Meta           map[string]any
}

// Error implements the error interface.
func (e *AppError) Error() string {
	var buffer bytes.Buffer
	buffer.Grow(200)
	buffer.WriteString("Error: ")
	buffer.WriteString(e.Message)
	buffer.WriteString(", InternalCode Code: ")
	buffer.WriteString(strconv.Itoa(e.InternalCode))
	buffer.WriteString(", ExternalCode Code: ")
	buffer.WriteString(strconv.Itoa(e.ExternalCode))
	buffer.WriteString(", LogDescription Code: ")
	buffer.Write(e.LogDescription)
	buffer.WriteString(", LogStack Code: ")
	buffer.Write(e.LogStack)
	return buffer.String()
}

// // AddStack sets the stack trace on the AppError instance and returns it.
// // This can be useful for providing additional context when an error occurs.
// func (e *AppError) AddStack() *AppError {
// 	e.Stack = Callers().Export()
// 	return e
// }

// AddDescription sets the description field of the AppError instance and returns it.
// This can be useful for providing additional context when an error occurs.
func (e *AppError) OverwriteDescription(description []byte) *AppError {
	e.LogDescription = description
	return e
}
func (e *AppError) AddDescription(description []byte) *AppError {
	e.LogDescription = append(e.LogDescription, description...)
	return e
}
func (e *AppError) OverwriteInternalCode(code int) *AppError {
	e.InternalCode = code
	return e
}
func (e *AppError) OverwriteExternalCode(code int) *AppError {
	e.ExternalCode = code
	return e
}

func (e *AppError) AddMeta(key string, value any) *AppError {
	if len(e.Meta) == 0 {
		e.Meta = make(map[string]any, 5)
	}
	e.Meta[key] = value
	return e
}

func NewAppError(status int, message string, description interface{}) *AppError {
	return &AppError{

		Message: message,
	}
}

func getErrorFromConfig(key string) *AppError {
	if err, exists := errorConfig[key]; exists {

		err.LogDescription = err.LogDescription
		return &err
	}
	return &AppError{

		ExternalCode: 500,
		Message:      "Unknown error",
	}
}
