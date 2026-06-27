package usecase

import (
	"time"

	"github.com/numduel/numduel/internal/domain"
)

// AuthDeps は認証系 UseCase の依存。
type AuthDeps struct {
	Repo                   domain.Repository
	Passwords              domain.PasswordHasher
	AccessTokens           domain.AccessTokenIssuer
	RefreshTokens          domain.RefreshTokenGenerator
	RefreshTokenExpiryDays int
	Now                    func() time.Time
}

func (d *AuthDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}
