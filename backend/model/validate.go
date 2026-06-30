package model

import (
	"regexp"
	"strings"
)

const (
	MinUsernameLength   = 3
	MaxUsernameLength   = 50
	MaxEmailLength      = 255
	MaxLoginEmailLength = 50
	MinPasswordLength   = 8
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateUsername(username string) error {
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength || !usernamePattern.MatchString(username) {
		return ErrBadUsername
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) > MaxEmailLength {
		return ErrBadEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

func ValidateLoginEmail(email string) error {
	if len(email) == 0 || len(email) > MaxLoginEmailLength {
		return ErrBadLoginEmail
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return ErrBadLoginEmail
	}
	return nil
}

func ValidateFourDigits(input string) ([4]int, error) {
	if len(input) != 4 {
		return [4]int{}, ErrBadDigitLength
	}
	seen := map[int]struct{}{}
	var digits [4]int
	for i, ch := range input {
		if ch < '0' || ch > '9' {
			return [4]int{}, ErrBadDigit
		}
		d := int(ch - '0')
		if _, ok := seen[d]; ok {
			return [4]int{}, ErrDuplicateDigit
		}
		seen[d] = struct{}{}
		digits[i] = d
	}
	return digits, nil
}

func ValidateFourDigitsArray(digits [4]int) error {
	seen := map[int]struct{}{}
	for i := 0; i < 4; i++ {
		if digits[i] < 0 || digits[i] > 9 {
			return ErrBadDigit
		}
		if _, ok := seen[digits[i]]; ok {
			return ErrDuplicateDigit
		}
		seen[digits[i]] = struct{}{}
	}
	return nil
}
