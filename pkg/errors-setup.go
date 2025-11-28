package pkg

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	// Load the error configuration from the JSON file
	err := LoadErrorConfig("errors.json")
	if err != nil {
		panic("Failed to load error configuration: " + err.Error())
	}
}

// internal templates loaded from JSON. Callers should use NewFromTemplate
// to obtain a mutable copy of a template; templates themselves remain
// isolated and immutable from external packages.
var errorTemplates map[ErrorCode]*AppError

// NewFromTemplate returns a new *AppError copied from the named template.
// The returned instance is safe for the caller to mutate (AppendStackLog,
// OverwriteDescription, AddMeta, etc.) without affecting the original
// templates loaded from the JSON file. If the template name is unknown,
// a default 500 Unknown error is returned.
func NewAppError(name ErrorCode) *AppError {
	if errorTemplates == nil {
		return &AppError{externalCode: 500, message: "error config not loaded"}
	}
	t, ok := errorTemplates[name]
	if !ok || t == nil {
		return &AppError{externalCode: 500, message: "Unknown error"}
	}
	return copyAppError(t)
}

// reverse map for parsing string -> ErrorCode
var errorCodesByName = func() map[string]ErrorCode {
	m := make(map[string]ErrorCode, len(ErrorNames))
	for k, v := range ErrorNames {
		m[v] = k
	}
	return m
}()

// ParseErrorCode converts a template name (string) to ErrorCode.
// Returns the code and true if found.
func ParseErrorCode(name string) (ErrorCode, bool) {
	c, ok := errorCodesByName[name]
	return c, ok
}

func LoadErrorConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// raw structure used only for JSON unmarshalling (exported fields so
	// encoding/json can populate them). We convert these into AppError values
	// with unexported fields so callers cannot mutate fields directly.
	type appErrorRaw struct {
		Message      string          `json:"message,omitempty"`
		Description  json.RawMessage `json:"description,omitempty"`
		Stack        json.RawMessage `json:"stack,omitempty"`
		InternalCode int             `json:"internal_code,omitempty"`
		ExternalCode int             `json:"external_code,omitempty"`
		Meta         map[string]any  `json:"meta,omitempty"`
		Level        string          `json:"level,omitempty"`
	}

	var raw map[string]appErrorRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	errorTemplates = make(map[ErrorCode]*AppError, len(raw))
	for k, v := range raw {
		var desc []byte
		if len(v.Description) > 0 {
			desc = make([]byte, len(v.Description))
			copy(desc, v.Description)
		}
		var stack []byte
		if len(v.Stack) > 0 {
			stack = make([]byte, len(v.Stack))
			copy(stack, v.Stack)
		}

		if code, ok := ParseErrorCode(k); ok {
			errorTemplates[code] = &AppError{
				message:        v.Message,
				logDescription: desc,
				logStack:       stack,
				internalCode:   v.InternalCode,
				externalCode:   v.ExternalCode,
				meta:           v.Meta,
			}

		}
	}
	return nil
}

// copyAppError performs a deep-ish copy of AppError: copies byte slices and
// maps so the returned value can be mutated independently of the template.
func copyAppError(src *AppError) *AppError {
	if src == nil {
		return nil
	}
	var desc []byte
	if len(src.logDescription) > 0 {
		desc = make([]byte, len(src.logDescription))
		copy(desc, src.logDescription)
	}
	var stack []byte
	if len(src.logStack) > 0 {
		stack = make([]byte, len(src.logStack))
		copy(stack, src.logStack)
	}
	var meta map[string]any
	if len(src.meta) > 0 {
		meta = make(map[string]any, len(src.meta))
		for k, v := range src.meta {
			meta[k] = v
		}
	}
	return &AppError{
		message:        src.message,
		logDescription: desc,
		logStack:       stack,
		internalCode:   src.internalCode,
		externalCode:   src.externalCode,
		meta:           meta,
	}
}

// AppError definition and methods

type AppError struct {
	message        string         `json:"message,omitempty"`
	logDescription []byte         `json:"description,omitempty"`
	logStack       []byte         `json:"stack,omitempty"`
	logLevel       logrus.Level   `json:"level,omitempty"`
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

// AddDescription / OverwriteDescription
func (e *AppError) OverwriteLevel(level logrus.Level) *AppError {
	e.logLevel = level
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

// Description returns a copy of the log description bytes (or nil).
func (e *AppError) DescriptionStr() string {
	if e == nil || len(e.logDescription) == 0 {
		return ""
	}
	return string(e.logDescription)
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

// Stack returns a copy of the stored stack log bytes (or nil).
func (e *AppError) StackStr() string {
	if e == nil || len(e.logStack) == 0 {
		return ""
	}
	return string(e.logStack)
}

// Stack returns a copy of the stored stack log bytes (or nil).
func (e *AppError) LevelStr() string {
	if e == nil {
		return ""
	}
	return e.logLevel.String()
}

// Stack returns a copy of the stored stack log bytes (or nil).
func (e *AppError) Level() logrus.Level {
	if e.logLevel != 0 {
		return e.logLevel
	} else {
		return logrus.ErrorLevel
	}
}

// InternalCode returns the internal code.
func (e *AppError) InternalCode() int {
	if e == nil {
		return 0
	}
	return e.internalCode
}

// InternalCode returns the internal code.
func (e *AppError) InternalCodeStr() string {
	if e == nil {
		return "0"
	}
	return strconv.Itoa(e.internalCode)
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

func (e *AppError) AppendStackLog(depth ...int) *AppError {
	if e == nil {
		return e
	}
	d := 1
	if len(depth) > 0 && depth[0] > 0 {
		d = depth[0]
	}
	pc, file, line, ok := runtime.Caller(d)
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
		e.logStack = append(e.logStack, ' ')
	}
	e.logStack = append(e.logStack, entry...)
	return e
}

func CustomAppError(status int, message string, description interface{}) *AppError {
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
