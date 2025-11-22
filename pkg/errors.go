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
	message        string         `json:"message,omitempty"`
	logDescription []byte         `json:"description,omitempty"`
	logStack       []byte         `json:"stack,omitempty"`
	internalCode   int            `json:"internal_code,omitempty"`
	externalCode   int            `json:"external_code,omitempty"`
	meta           map[string]any `json:"meta,omitempty"`
}

// Error implements the error interface. nil-safe.
func (e *AppError) Error() string {
	if e == nil {
		return "<nil AppError>"
	}
	var buffer bytes.Buffer
	buffer.Grow(200)
	buffer.WriteString("Error: ")
	buffer.WriteString(e.message)
	buffer.WriteString(", InternalCode Code: ")
	buffer.WriteString(strconv.Itoa(e.internalCode))
	buffer.WriteString(", ExternalCode Code: ")
	buffer.WriteString(strconv.Itoa(e.externalCode))
	buffer.WriteString(", LogDescription Code: ")
	buffer.Write(e.logDescription)
	buffer.WriteString(", LogStack Code: ")
	buffer.Write(e.logStack)
	return buffer.String()
}

// AddDescription / OverwriteDescription
func (e *AppError) OverwriteDescription(description []byte) *AppError {
	e.logDescription = description
	return e
}
func (e *AppError) AddDescription(description []byte) *AppError {
	e.logDescription = append(e.logDescription, description...)
	return e
}
func (e *AppError) OverwriteInternalCode(code int) *AppError {
	e.internalCode = code
	return e
}
func (e *AppError) OverwriteExternalCode(code int) *AppError {
	e.externalCode = code
	return e
}

func (e *AppError) AddMeta(key string, value any) *AppError {
	if len(e.meta) == 0 {
		e.meta = make(map[string]any, 5)
	}
	e.meta[key] = value
	return e
}

func (e *AppError) AppendStackLog() *AppError {
// TODO
// must stack from run time get ,  and just 1 dept , for example evry where that AppendStackLog call , add same line

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
		message:        message,
		externalCode:   status,
		logDescription: descBytes,
	}
}

func getErrorFromConfig(key string) *AppError {
	if errorConfig == nil {
		return &AppError{
			externalCode: 500,
			message:      "error config not loaded",
		}
	}
	if errPtr, exists := errorConfig[key]; exists && errPtr != nil {
		return errPtr
	}
	return &AppError{
		externalCode: 500,
		message:      "Unknown error",
	}
}
