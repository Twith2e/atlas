package phone

import (
	appErr "atlas/internal/errors"
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D`)

func NormalizeNigerianNumber(phone string) (string, error) {
	cleaned := nonDigit.ReplaceAllString(strings.TrimSpace(phone), "")
	if cleaned == "" {
		return "", appErr.ErrInvalidPhoneNumber
	}

	switch {
	case strings.HasPrefix(cleaned, "0") && len(cleaned) == 11:
		return "234" + cleaned[1:], nil
	case strings.HasPrefix(cleaned, "234") && len(cleaned) == 13:
		return cleaned, nil
	case len(cleaned) == 10:
		return "234" + cleaned, nil
	}
	return "", appErr.ErrInvalidPhoneNumber
}
