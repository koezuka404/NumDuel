package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type User struct {
	ID             uuid.UUID
	Username       string
	Email          string
	PasswordHash   string
	Role           Role
	WinCount       int
	DeletedAt      *time.Time
	DeletedBy      *uuid.UUID
	LastActivityAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) IsDeleted() bool {
	return u != nil && u.DeletedAt != nil
}

func (u *User) IsMaster() bool {
	return u != nil && u.Role == RoleMaster
}

func (u *User) CanMatch() bool {
	return u != nil && !u.IsDeleted() && !u.IsMaster()
}

func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 || !usernamePattern.MatchString(username) {
		return errValidation("username must be 3-50 alphanumeric/underscore characters")
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) > 255 {
		return errValidation("email must be 1-255 characters")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errValidation("password must be at least 8 characters")
	}
	return nil
}

func ValidateLoginEmail(email string) error {
	if len(email) == 0 || len(email) > 50 {
		return errValidation("email must be 1-50 characters")
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return errValidation("email format is invalid")
	}
	return nil
}
