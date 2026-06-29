// 認証系 UseCase。ビジネスロジックと DB トランザクションを担当。
package usecase

import (
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// AuthDeps は認証 UseCase が使う依存関係の集合。
type AuthDeps struct {
	Repo                   repository.IRepository
	Tx                     repository.TxManager
	Passwords              model.PasswordHasher
	AccessTokens           model.AccessTokenIssuer
	RefreshTokens          model.RefreshTokenGenerator
	JWTRevoker             model.JWTRevoker     // ログアウト時の JWT 失効
	WSSessions             model.WSSessionStore // ログアウト時の WS 切断
	RefreshTokenExpiryDays int
	Now                    func() time.Time // テスト用。nil なら time.Now().UTC()
}

func (d AuthDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}
