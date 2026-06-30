// User エンティティと入力バリデーション
package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Username       string     `gorm:"size:50;not null;uniqueIndex"`
	Email          string     `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash   string     `gorm:"size:255;not null;column:password"`
	Role           Role       `gorm:"size:20;not null"`
	WinCount       int        `gorm:"not null;default:0"`
	DeletedAt      *time.Time `gorm:"index"`
	DeletedBy      *uuid.UUID `gorm:"type:uuid"`
	LastActivityAt time.Time  `gorm:"not null"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (User) TableName() string { return "users" }

func (u *User) IsDeleted() bool {
	return u != nil && u.DeletedAt != nil
}

func (u *User) IsMaster() bool {
	return u != nil && u.Role == RoleMaster
}

// CanMatch はマッチング参加可能か（削除済み・master は不可）
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
