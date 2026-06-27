package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/infrastructure/crypto"
)

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
	hash, err := crypto.NewPasswordService().Hash(password)
	if err != nil {
		return fmt.Errorf("hash master password: %w", err)
	}
	now := time.Now().UTC()
	return repo.Users().Create(ctx, nil, &domain.User{
		ID: uuid.New(), Username: masterUsername(email), Email: email, PasswordHash: hash,
		Role: domain.RoleMaster, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	})
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
