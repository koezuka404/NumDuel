package usecase

import (
	"regexp"
	"strings"

	"github.com/numduel/numduel/model"
)

const (
	minUsernameLength   = 3
	maxUsernameLength   = 50
	maxEmailLength      = 255
	maxLoginEmailLength = 50
	minPasswordLength   = 8
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateUsername(username string) error {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength || !usernamePattern.MatchString(username) {
		return model.ErrBadUsername
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) > maxEmailLength {
		return model.ErrBadEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return model.ErrWeakPassword
	}
	return nil
}

func ValidateLoginEmail(email string) error {
	if len(email) == 0 || len(email) > maxLoginEmailLength {
		return model.ErrBadLoginEmail
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return model.ErrBadLoginEmail
	}
	return nil
}
