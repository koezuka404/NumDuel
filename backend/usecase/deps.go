// 認証系 UseCase。ビジネスロジックと DB トランザクションを担当。
package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/model"
)

// AuthDeps は認証 UseCase が使う依存関係の集合。
type AuthDeps struct {
	Repo                   model.Repository
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

// withTx はトランザクションを開始し、fn 成功時のみコミットする。
func withTx(ctx context.Context, repo model.Repository, fn func(model.Transaction) error) error {
	tx, err := repo.Begin(ctx)
	if err != nil {
		return model.ErrInternal("failed to begin transaction")
	}
	defer func() { _ = repo.Rollback(tx) }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := repo.Commit(tx); err != nil {
		return model.ErrInternal("failed to commit transaction")
	}
	return nil
}
