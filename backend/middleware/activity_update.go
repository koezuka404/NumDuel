//JWT必須APIでusers.last_activity_atを更新する（AutoLogoutWorker用）
package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/repository"
)

//ActivityUpdateConfigはActivityUpdateの依存関係
//Auth通過後のprotectedルートでのみlast_activity_atを更新する
type ActivityUpdateConfig struct {
	Repo repository.Repos
}

//ActivityUpdateは認証済みリクエスト処理後にlast_activity_atをnowに更新する
//AutoLogoutWorkerがSESSION_TIMEOUT_MINUTES超過を判定するためHTTP操作もセッション継続扱いにする
//login/refresh/logout/WSは各UseCase・Handler側でも更新する
//更新失敗時もハンドラのレスポンスはそのまま返す
func ActivityUpdate(cfg ActivityUpdateConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			//AuthMiddlewareがSetAuthした場合のみ更新（register/login/refreshは対象外）
			auth, ok := AuthFrom(c)
			if !ok || cfg.Repo.User == nil {
				return err
			}
			now := time.Now().UTC()
			_ = cfg.Repo.User.TouchLastActivity(c.Request().Context(), auth.UserID, now)
			return err
		}
	}
}
