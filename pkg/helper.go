package pkg

import (
	"os"
	"time"
)

func PtrString(s string) *string {
	return &s
}

func PtrInt(i int) *int {
	return &i
}

func PtrUint64(i uint64) *uint64 {
	return &i
}

func PtrBool(b bool) *bool {
	return &b
}

func PtrTime(t time.Time) *time.Time {
	return &t
}

func EnvOrDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return def
	}
	return val
}
