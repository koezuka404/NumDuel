package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/numduel/numduel/internal/domain"
)

const bcryptCost = 12

// SeedMaster は users に master が 0 件のときのみ初回 master を作成する。
func SeedMaster(ctx context.Context, repo domain.Repository, email, password string) error {
	count, err := repo.Users().CountMasters(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if email == "" || password == "" {
		return fmt.Errorf("NUMDUEL_MASTER_EMAIL and NUMDUEL_MASTER_PASSWORD are required when no master exists")
	}
	if len(password) < 8 {
		return fmt.Errorf("NUMDUEL_MASTER_PASSWORD must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash master password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:             uuid.New(),
		Username:       masterUsername(email),
		Email:          email,
		PasswordHash:   string(hash),
		Role:           domain.RoleMaster,
		WinCount:       0,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return repo.Users().Create(ctx, nil, user)
}

func masterUsername(email string) string {
	local := strings.Split(email, "@")[0]
	local = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, local)
	if len(local) < 3 {
		return "master"
	}
	if len(local) > 50 {
		return local[:50]
	}
	return local
}
