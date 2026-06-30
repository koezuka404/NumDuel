// JWT 必須 API で users.last_activity_at を更新する（AutoLogoutWorker 用）
package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/repository"
)

// ActivityUpdateConfig は ActivityUpdate の依存関係
// Auth 通過後の protected ルートでのみ last_activity_at を更新する
type ActivityUpdateConfig struct {
	Repo repository.Repos
}

// ActivityUpdate は認証済みリクエスト処理後に last_activity_at を now に更新する
// AutoLogoutWorker が SESSION_TIMEOUT_MINUTES 超過を判定するため HTTP 操作もセッション継続扱いにする
// login / refresh / logout / WS は各 UseCase・Handler 側でも更新する
// 更新失敗時もハンドラのレスポンスはそのまま返す
func ActivityUpdate(cfg ActivityUpdateConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			// Auth Middleware が SetAuth した場合のみ更新（register/login/refresh は対象外）
			auth, ok := AuthFrom(c)
			if !ok || cfg.Repo == nil {
				return err
			}
			now := time.Now().UTC()
			_ = cfg.Repo.User.TouchLastActivity(c.Request().Context(), auth.UserID, now)
			return err
		}
	}
}
