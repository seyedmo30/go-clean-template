package pkg

import "time"

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
