package pkg

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"
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

// Getters (nil-safe). For slices/maps we return copies to avoid accidental
// external mutation of internal state and to be safe for concurrent readers.

// Message returns the error message.
func (e *AppError) Message() string {
	if e == nil {
		return ""
	}
	return e.message
}

// Description returns a copy of the log description bytes (or nil).
func (e *AppError) Description() []byte {
	if e == nil || len(e.logDescription) == 0 {
		return nil
	}
	b := make([]byte, len(e.logDescription))
	copy(b, e.logDescription)
	return b
}

// Stack returns a copy of the stored stack log bytes (or nil).
func (e *AppError) Stack() []byte {
	if e == nil || len(e.logStack) == 0 {
		return nil
	}
	b := make([]byte, len(e.logStack))
	copy(b, e.logStack)
	return b
}

// InternalCode returns the internal code.
func (e *AppError) InternalCode() int {
	if e == nil {
		return 0
	}
	return e.internalCode
}

// ExternalCode returns the external (status) code.
func (e *AppError) ExternalCode() int {
	if e == nil {
		return 0
	}
	return e.externalCode
}

// Meta returns a shallow copy of the meta map (or nil).
func (e *AppError) Meta() map[string]any {
	if e == nil || len(e.meta) == 0 {
		return nil
	}
	m := make(map[string]any, len(e.meta))
	for k, v := range e.meta {
		m[k] = v
	}
	return m
}

// GetMeta retrieves a single meta value by key.
func (e *AppError) GetMeta(key string) (any, bool) {
	if e == nil {
		return nil, false
	}
	v, ok := e.meta[key]
	return v, ok
}




func (e *AppError) AppendStackLog() *AppError {
	if e == nil {
		return e
	}
	pc, file, line, ok := runtime.Caller(1)
	var entry []byte
	if ok {
		// find last path separator ('/' or '\') and slice base name
		idx := strings.LastIndexAny(file, "/\\")
		base := file
		if idx != -1 {
			base = file[idx+1:]
		}

		entry = append(entry, base...)
		entry = append(entry, ':')
		entry = strconv.AppendInt(entry, int64(line), 10)
		entry = append(entry, ' ')
		if fn := runtime.FuncForPC(pc); fn != nil {
			entry = append(entry, fn.Name()...)
		} else {
			entry = append(entry, "unknown"...)
		}
	} else {
		entry = append(entry, "unknown"...)
	}

	if len(e.logStack) > 0 {
		e.logStack = append(e.logStack, '\n')
	}
	e.logStack = append(e.logStack, entry...)
	return e
}

