// JWT 認証 Middleware。Controller の前段でトークンを検証する。
package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	infrcrypto "github.com/numduel/numduel/internal/infrastructure/crypto"
	"github.com/numduel/numduel/internal/httputil"
)

// Auth は Authorization: Bearer ヘッダーを検証し、ユーザー情報をコンテキストに保存する。
func Auth(jwt *infrcrypto.JWTService, revoker domain.JWTRevoker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return httputil.WriteError(c, domain.ErrUnauthorized())
			}
			token, err := jwt.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				return httputil.WriteError(c, err)
			}
			// ログアウト済み JWT の再利用を防ぐ（Redis 実装後に有効）
			if revoker != nil {
				revoked, err := revoker.IsRevoked(c.Request().Context(), token.JTI)
				if err != nil {
					return httputil.WriteError(c, domain.ErrInternal("failed to check token revocation"))
				}
				if revoked {
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
