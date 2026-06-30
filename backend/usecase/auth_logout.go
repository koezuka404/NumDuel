package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type LogoutInput struct {
	UserID uuid.UUID
	JTI    string
	Exp    time.Time
}

func (a *AuthUseCase) Logout(ctx context.Context, in LogoutInput) error {
	now := a.now()
	if in.JTI == "" || !now.Before(in.Exp) {
		return ErrUnauthorized
	}
	if a.JWTRevoker != nil {
		if err := a.JWTRevoker.Revoke(ctx, in.JTI, in.Exp.Sub(now)); err != nil {
			return err
		}
	}
	if a.WSSessions != nil {
		_ = a.WSSessions.DeleteUser(ctx, in.UserID)
	}
	return repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		if err := revokeRefreshTokensByUserID(ctx, a.RefreshTokens, in.UserID, now); err != nil {
			return err
		}
		if err := a.createLoginLog(ctx, in.UserID, model.LoginActionLogout, now); err != nil {
			return err
		}
		user, err := a.Users.FindByID(ctx, in.UserID)
		if err != nil {
			return err
		}
		if user != nil {
			return a.touchUserActivity(ctx, user, now)
		}
		return nil
	})
}
