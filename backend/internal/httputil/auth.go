// Middleware が検証した JWT 情報を Controller へ渡すためのコンテキスト保持。
package httputil

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
)

// AuthInfo は JWT から取り出した認証情報。
type AuthInfo struct {
	UserID    uuid.UUID
	Role      domain.Role
	JTI       string // JWT ID（失効管理用）
	ExpiresAt time.Time
}

const authKey = "auth"

func SetAuth(c echo.Context, info AuthInfo) { c.Set(authKey, info) }

func AuthFrom(c echo.Context) (AuthInfo, bool) {
	v, ok := c.Get(authKey).(AuthInfo)
	return v, ok
}
