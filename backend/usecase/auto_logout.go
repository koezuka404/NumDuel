package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// AutoLogoutDeps は AutoLogoutUseCase の依存関係
// SESSION_TIMEOUT_MINUTES 分間 last_activity_at が更新されていないユーザーを対象にする
type AutoLogoutDeps struct {
	Repo            repository.Repos
	ForceLogout     model.IForceLogoutStore // Redis user:{userId}:force_logout_before
	ForceDisconnect func(ctx context.Context, userID uuid.UUID) error // WS ERROR 送信後に切断
	SessionTimeout  time.Duration
	Now             func() time.Time
}

func (d AutoLogoutDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

// AutoLogout は無操作ユーザーを強制ログアウトする
// last_activity_at < now - SessionTimeout のユーザーを列挙し、1 件ずつ autoLogoutUser する
func AutoLogout(ctx context.Context, d AutoLogoutDeps) error {
	if d.SessionTimeout <= 0 {
		return nil
	}
	now := d.now()
	before := now.Add(-d.SessionTimeout)
	users, err := d.Repo.User.ListInactiveSince(ctx, before)
	if err != nil {
		return model.ErrInternal("failed to list inactive users")
	}
	for _, user := range users {
		if user == nil {
			continue
		}
		if err := autoLogoutUser(ctx, d, user.ID, now); err != nil {
			return err
		}
	}
	return nil
}

// autoLogoutUser は 1 ユーザーのセッションを切る
// 1) force_logout_before SET  2) WS 切断  3) refresh 失効 + login_logs(auto_logout) + last_activity_at 更新
func autoLogoutUser(ctx context.Context, d AutoLogoutDeps, userID uuid.UUID, now time.Time) error {
	if d.ForceLogout != nil {
		if err := d.ForceLogout.SetForceLogoutBefore(ctx, userID, now); err != nil {
			return model.ErrInternal("failed to set force logout")
		}
	}
	if d.ForceDisconnect != nil {
		_ = d.ForceDisconnect(ctx, userID)
	}
	return repository.WithTx(ctx, d.Repo.DB, func(ctx context.Context) error {
		if err := revokeRefreshTokensByUserID(ctx, d.Repo, userID, now); err != nil {
			return model.ErrInternal("failed to revoke refresh tokens")
		}
		if err := d.Repo.LoginLog.Create(ctx, &model.LoginLog{
			ID: uuid.New(), UserID: userID, Action: model.LoginActionAutoLogout,
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return model.ErrInternal("failed to create login log")
		}
		user, err := d.Repo.User.FindByID(ctx, userID)
		if err != nil {
			return model.ErrInternal("failed to find user")
		}
		if user != nil && !user.IsDeleted() {
			user.LastActivityAt = now
			user.UpdatedAt = now
			if err := d.Repo.User.Update(ctx, user); err != nil {
				return model.ErrInternal("failed to update user")
			}
		}
		return nil
	})
}
