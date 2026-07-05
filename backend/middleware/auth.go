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
			info, err := resolveAuthFromCookie(c, cfg, true)
			if err != nil {
				return dto.WriteError(c, err)
			}
			if info.UserID == uuid.Nil {
				return dto.WriteError(c, usecase.ErrUnauthorized)
			}
			SetAuth(c, info)
			return next(c)
		}
	}
}

//TryAuthはCookieが有効なときだけAuthInfoをセットし、未ログインでも401にしない
func TryAuth(cfg AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			info, err := resolveAuthFromCookie(c, cfg, false)
			if err == nil && info.UserID != uuid.Nil {
				SetAuth(c, info)
			}
			return next(c)
		}
	}
}

func resolveAuthFromCookie(c echo.Context, cfg AuthConfig, strict bool) (AuthInfo, error) {
	cookie, err := c.Cookie(AccessCookieName)
	if err != nil || cookie.Value == "" {
		if strict {
			return AuthInfo{}, usecase.ErrUnauthorized
		}
		return AuthInfo{}, nil
	}
	token, err := cfg.JWT.Parse(cookie.Value)
	if err != nil {
		if strict {
			return AuthInfo{}, err
		}
		return AuthInfo{}, nil
	}
	if cfg.Revoker != nil {
		revoked, err := authCheckRevoked(c.Request().Context(), cfg.Revoker, token.JTI)
		if err != nil {
			if strict {
				return AuthInfo{}, err
			}
			return AuthInfo{}, nil
		}
		if revoked {
			if strict {
				return AuthInfo{}, usecase.ErrUnauthorized
			}
			return AuthInfo{}, nil
		}
	}
	if cfg.ForceLogout != nil && !token.IssuedAt.IsZero() {
		before, err := authForceLogoutBefore(c.Request().Context(), cfg.ForceLogout, token.UserID)
		if err != nil {
			if strict {
				return AuthInfo{}, err
			}
			return AuthInfo{}, nil
		}
		if !before.IsZero() && token.IssuedAt.Before(before) {
			if strict {
				return AuthInfo{}, usecase.ErrUnauthorized
			}
			return AuthInfo{}, nil
		}
	}
	if cfg.Repo.User != nil {
		user, err := cfg.Repo.User.FindByID(c.Request().Context(), token.UserID)
		if err != nil {
			if strict {
				return AuthInfo{}, err
			}
			return AuthInfo{}, nil
		}
		if user == nil || user.IsDeleted() {
			if strict {
				return AuthInfo{}, usecase.ErrUnauthorized
			}
			return AuthInfo{}, nil
		}
	}
	return AuthInfo{
		UserID: token.UserID, Role: token.Role, JTI: token.JTI, ExpiresAt: token.ExpiresAt,
	}, nil
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
