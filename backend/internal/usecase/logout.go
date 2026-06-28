package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

type LogoutInput struct {
	UserID uuid.UUID
	JTI    string    // JWT ID
	Exp    time.Time // JWT 有効期限（Redis TTL 計算用）
}

// Logout は JWT 失効、refresh 全失効、WS 切断、login_logs を記録する。
func Logout(ctx context.Context, d AuthDeps, in LogoutInput) error {
	now := d.now()
	if in.JTI == "" || !now.Before(in.Exp) {
		return domain.ErrUnauthorized()
	}
	if d.JWTRevoker != nil {
		if err := d.JWTRevoker.Revoke(ctx, in.JTI, in.Exp.Sub(now)); err != nil {
			return domain.ErrInternal("failed to revoke jwt")
		}
	}
	if d.WSSessions != nil {
		_ = d.WSSessions.DeleteUser(ctx, in.UserID)
	}
	return withTx(ctx, d.Repo, func(tx domain.Transaction) error {
		if err := d.Repo.RefreshTokens().RevokeAllActiveByUserID(ctx, tx, in.UserID, now); err != nil {
			return domain.ErrInternal("failed to revoke refresh tokens")
		}
		if err := d.Repo.LoginLogs().Create(ctx, tx, &domain.LoginLog{
			ID: uuid.New(), UserID: in.UserID, Action: domain.LoginActionLogout, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return domain.ErrInternal("failed to create login log")
		}
		user, err := d.Repo.Users().FindByID(ctx, in.UserID)
		if err != nil {
			return domain.ErrInternal("failed to find user")
		}
		if user != nil {
			user.LastActivityAt = now
			user.UpdatedAt = now
			if err := d.Repo.Users().Update(ctx, tx, user); err != nil {
				return domain.ErrInternal("failed to update user activity")
			}
		}
		return nil
	})
}
