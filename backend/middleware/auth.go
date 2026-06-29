// JWT 認証 Middleware。Cookie access_token を検証する。
package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/dto"
)

type AuthConfig struct {
	JWT         *infrcrypto.JWTService
	Revoker     model.JWTRevoker
	ForceLogout model.ForceLogoutStore
	Repo        model.Repository
}

// Auth は HttpOnly Cookie access_token を検証し、ユーザー情報をコンテキストに保存する。
func Auth(cfg AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(AccessCookieName)
			if err != nil || cookie.Value == "" {
				return dto.WriteError(c, model.ErrUnauthorized())
			}
			token, err := cfg.JWT.Parse(cookie.Value)
			if err != nil {
				return dto.WriteError(c, err)
			}
			if cfg.Revoker != nil {
				revoked, err := cfg.Revoker.IsRevoked(c.Request().Context(), token.JTI)
				if err != nil {
					return dto.WriteError(c, model.ErrInternal("failed to check token revocation"))
				}
				if revoked {
					return dto.WriteError(c, model.ErrUnauthorized())
				}
			}
			if cfg.ForceLogout != nil && !token.IssuedAt.IsZero() {
				before, err := cfg.ForceLogout.GetForceLogoutBefore(c.Request().Context(), token.UserID)
				if err != nil {
					return dto.WriteError(c, model.ErrInternal("failed to check force logout"))
				}
				if !before.IsZero() && token.IssuedAt.Before(before) {
					return dto.WriteError(c, model.ErrUnauthorized())
				}
			}
			if cfg.Repo != nil {
				user, err := cfg.Repo.Users().FindByID(c.Request().Context(), token.UserID)
				if err != nil {
					return dto.WriteError(c, model.ErrInternal("failed to find user"))
				}
				if user == nil || user.IsDeleted() {
					return dto.WriteError(c, model.ErrUnauthorized())
				}
			}
			SetAuth(c, AuthInfo{
				UserID: token.UserID, Role: token.Role, JTI: token.JTI, ExpiresAt: token.ExpiresAt,
			})
			return next(c)
		}
	}
}
