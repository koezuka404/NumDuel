//JWT認証MiddlewareCookieaccess_tokenを検証する
package middleware

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/usecase"
)

type AuthConfig struct {
	JWT         *infrcrypto.JWTService
	Revoker     usecase.IJWTRevoker
	ForceLogout usecase.IForceLogoutStore
	Repo        repository.Repos
}

func Auth(cfg AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(AccessCookieName)
			if err != nil || cookie.Value == "" {
				return dto.WriteError(c, usecase.ErrUnauthorized)
			}
			token, err := cfg.JWT.Parse(cookie.Value)
			if err != nil {
				return dto.WriteError(c, err)
			}
			if cfg.Revoker != nil {
				revoked, err := authCheckRevoked(c.Request().Context(), cfg.Revoker, token.JTI)
				if err != nil {
					return dto.WriteError(c, err)
				}
				if revoked {
					return dto.WriteError(c, usecase.ErrUnauthorized)
				}
			}
			if cfg.ForceLogout != nil && !token.IssuedAt.IsZero() {
				before, err := authForceLogoutBefore(c.Request().Context(), cfg.ForceLogout, token.UserID)
				if err != nil {
					return dto.WriteError(c, err)
				}
				if !before.IsZero() && token.IssuedAt.Before(before) {
					return dto.WriteError(c, usecase.ErrUnauthorized)
				}
			}
			if cfg.Repo.User != nil {
				user, err := cfg.Repo.User.FindByID(c.Request().Context(), token.UserID)
				if err != nil {
					return dto.WriteError(c, err)
				}
				if user == nil || user.IsDeleted() {
					return dto.WriteError(c, usecase.ErrUnauthorized)
				}
			}
			SetAuth(c, AuthInfo{
				UserID: token.UserID, Role: token.Role, JTI: token.JTI, ExpiresAt: token.ExpiresAt,
			})
			return next(c)
		}
	}
}

func authCheckRevoked(ctx context.Context, revoker usecase.IJWTRevoker, jti string) (bool, error) {
	revoked, err := revoker.IsRevoked(ctx, jti)
	if err != nil {
		log.Printf("auth: IsRevoked degraded (redis unavailable): %v", err)
		return false, nil
	}
	return revoked, nil
}

func authForceLogoutBefore(ctx context.Context, store usecase.IForceLogoutStore, userID uuid.UUID) (time.Time, error) {
	before, err := store.GetForceLogoutBefore(ctx, userID)
	if err != nil {
		log.Printf("auth: GetForceLogoutBefore degraded (redis unavailable): %v", err)
		return time.Time{}, nil
	}
	return before, nil
}
