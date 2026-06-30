package model

import (
	"time"

	"github.com/google/uuid"
)

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

func (u User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u User) IsMaster() bool {
	return u.Role == RoleMaster
}

func (u User) CanMatch() bool {
	return !u.IsDeleted() && !u.IsMaster()
}
