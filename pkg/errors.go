package pkg

type ErrorCode int

const (
	ErrBadRequest ErrorCode = iota
	ErrNotFound

// add more error codes as needed
)

// ErrorNames maps ErrorCode values to template keys in the JSON file.
var ErrorNames = map[ErrorCode]string{
	ErrBadRequest: "ErrBadRequest",
	ErrNotFound:   "ErrNotFound",
}

// String implements fmt.Stringer.
func (c ErrorCode) String() string {
	if s, ok := ErrorNames[c]; ok {
		return s
	}
	return "UnknownErrorCode"
}
