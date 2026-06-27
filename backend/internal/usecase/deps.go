package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/internal/domain"
)

type AuthDeps struct {
	Repo                   domain.Repository
	Passwords              domain.PasswordHasher
	AccessTokens           domain.AccessTokenIssuer
	RefreshTokens          domain.RefreshTokenGenerator
	RefreshTokenExpiryDays int
	Now                    func() time.Time
}

func (d AuthDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

func withTx(ctx context.Context, repo domain.Repository, fn func(domain.Transaction) error) error {
	tx, err := repo.Begin(ctx)
	if err != nil {
		return domain.ErrInternal("failed to begin transaction")
	}
	defer func() { _ = repo.Rollback(tx) }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := repo.Commit(tx); err != nil {
		return domain.ErrInternal("failed to commit transaction")
	}
	return nil
}
