// JWT 認証 Middleware。Controller の前段でトークンを検証する。
package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	infrcrypto "github.com/numduel/numduel/internal/infrastructure/crypto"
	"github.com/numduel/numduel/internal/httputil"
)

type AuthConfig struct {
	JWT         *infrcrypto.JWTService
	Revoker     domain.JWTRevoker
	ForceLogout domain.ForceLogoutStore
	Repo        domain.Repository
}

// Auth は Authorization: Bearer ヘッダーを検証し、ユーザー情報をコンテキストに保存する。
func Auth(cfg AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return httputil.WriteError(c, domain.ErrUnauthorized())
			}
			token, err := cfg.JWT.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				return httputil.WriteError(c, err)
			}
			if cfg.Revoker != nil {
				revoked, err := cfg.Revoker.IsRevoked(c.Request().Context(), token.JTI)
				if err != nil {
					return httputil.WriteError(c, domain.ErrInternal("failed to check token revocation"))
				}
				if revoked {
					return httputil.WriteError(c, domain.ErrUnauthorized())
				}
			}
			if cfg.ForceLogout != nil && !token.IssuedAt.IsZero() {
				before, err := cfg.ForceLogout.GetForceLogoutBefore(c.Request().Context(), token.UserID)
				if err != nil {
					return httputil.WriteError(c, domain.ErrInternal("failed to check force logout"))
				}
				if !before.IsZero() && token.IssuedAt.Before(before) {
					return httputil.WriteError(c, domain.ErrUnauthorized())
				}
			}
			if cfg.Repo != nil {
				user, err := cfg.Repo.Users().FindByID(c.Request().Context(), token.UserID)
				if err != nil {
					return httputil.WriteError(c, domain.ErrInternal("failed to find user"))
				}
				if user == nil || user.IsDeleted() {
					return httputil.WriteError(c, domain.ErrUnauthorized())
				}
			}
			httputil.SetAuth(c, httputil.AuthInfo{
				UserID: token.UserID, Role: token.Role, JTI: token.JTI, ExpiresAt: token.ExpiresAt,
			})
			return next(c)
		}
	}
}
