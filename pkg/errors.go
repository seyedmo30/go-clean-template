package pkg

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
)

func init() {
	// Load the error configuration from the JSON file
	err := LoadErrorConfig("errors.json")
	if err != nil {
		panic("Failed to load error configuration: " + err.Error())
	}
}

var errorConfig map[string]*AppError

var (
	ErrBadRequest *AppError
	ErrNotFound   *AppError
	// TODO some common error
)

func LoadErrorConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	// unmarshal into map of pointers to avoid copying issues
	err = json.Unmarshal(data, &errorConfig)
	if err != nil {
		return err
	}
	initializeErrors()
	return nil
}

func initializeErrors() {
	ErrBadRequest = getErrorFromConfig("ErrBadRequest")
	ErrNotFound = getErrorFromConfig("ErrNotFound")
}

type AppError struct {
	Message        string         `json:"message,omitempty"`
	LogDescription []byte         `json:"description,omitempty"`
	LogStack       []byte         `json:"stack,omitempty"`
	InternalCode   int            `json:"internal_code,omitempty"`
	ExternalCode   int            `json:"external_code,omitempty"`
	Meta           map[string]any `json:"meta,omitempty"`
}

// Error implements the error interface. nil-safe.
func (e *AppError) Error() string {
	if e == nil {
		return "<nil AppError>"
	}
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

// AddDescription / OverwriteDescription
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

// NewAppError creates an AppError; status is used as ExternalCode.
func NewAppError(status int, message string, description interface{}) *AppError {
	var descBytes []byte
	switch v := description.(type) {
	case nil:
		// nothing
	case string:
		descBytes = []byte(v)
	case []byte:
		descBytes = v
	default:
		// try to JSON-encode other types
		if b, err := json.Marshal(v); err == nil {
			descBytes = b
		}
	}
	return &AppError{
		Message:        message,
		ExternalCode:   status,
		LogDescription: descBytes,
	}
}

func getErrorFromConfig(key string) *AppError {
	if errorConfig == nil {
		return &AppError{
			ExternalCode: 500,
			Message:      "error config not loaded",
		}
	}
	if errPtr, exists := errorConfig[key]; exists && errPtr != nil {
		return errPtr
	}
	return &AppError{
		ExternalCode: 500,
		Message:      "Unknown error",
	}
}
