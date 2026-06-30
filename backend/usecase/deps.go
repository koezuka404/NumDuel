// 認証系 UseCaseビジネスロジックと DB トランザクションを担当
package usecase

import (
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// AuthDeps は認証 UseCase が使う依存関係の集合
type AuthDeps struct {
	Repo                   repository.Repos
	Passwords              model.IPasswordHasher
	AccessTokens           model.IAccessTokenIssuer
	RefreshTokens          model.IRefreshTokenGenerator
	JWTRevoker             model.IJWTRevoker     // ログアウト時の JWT 失効
	WSSessions             model.IWSSessionStore // ログアウト時の WS 切断
	RefreshTokenExpiryDays int
	Now                    func() time.Time // テスト用nil なら time.Now().UTC()
}

func (d AuthDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}
