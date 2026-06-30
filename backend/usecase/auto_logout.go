package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// AutoLogoutDeps は AutoLogoutUseCase の依存関係
type AutoLogoutDeps struct {
	Repo            repository.IRepository
	Tx              repository.TxManager
	ForceLogout     model.ForceLogoutStore
	ForceDisconnect func(ctx context.Context, userID uuid.UUID) error
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
func AutoLogout(ctx context.Context, d AutoLogoutDeps) error {
	if d.SessionTimeout <= 0 {
		return nil
	}
	now := d.now()
	before := now.Add(-d.SessionTimeout)
	users, err := d.Repo.Users().ListInactiveSince(ctx, before)
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

func autoLogoutUser(ctx context.Context, d AutoLogoutDeps, userID uuid.UUID, now time.Time) error {
	if d.ForceLogout != nil {
		if err := d.ForceLogout.SetForceLogoutBefore(ctx, userID, now); err != nil {
			return model.ErrInternal("failed to set force logout")
		}
	}
	if d.ForceDisconnect != nil {
		_ = d.ForceDisconnect(ctx, userID)
	}
	return d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		if err := revokeRefreshTokensByUserID(ctx, tx, userID, now); err != nil {
			return model.ErrInternal("failed to revoke refresh tokens")
		}
		if err := tx.LoginLogs().Create(ctx, &model.LoginLog{
			ID: uuid.New(), UserID: userID, Action: model.LoginActionAutoLogout,
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return model.ErrInternal("failed to create login log")
		}
		user, err := tx.Users().FindByID(ctx, userID)
		if err != nil {
			return model.ErrInternal("failed to find user")
		}
		if user != nil && !user.IsDeleted() {
			user.LastActivityAt = now
			user.UpdatedAt = now
			if err := tx.Users().Update(ctx, user); err != nil {
				return model.ErrInternal("failed to update user")
			}
		}
		return nil
	})
}
