package usecase

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/entity/user"
	"strings"
	"unicode/utf8"
)

func UserIntegrationValidate(req *integration.UserDTO) error {
	const maxPhoneLen = 20
	// convert named string type to builtin string for processing
	phoneStr := strings.TrimSpace(string(req.Phone))

	// rune-safe truncate
	if utf8.RuneCountInString(phoneStr) > maxPhoneLen {
		runes := []rune(phoneStr)
		phoneStr = string(runes[:maxPhoneLen])
	}

	// convert back to the named Phone type and write back
	req.Phone = user.Phone(phoneStr)

	return nil
}
