package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type WSAuthDeps struct {
	Repo        repository.Repos
	JWT         *infrcrypto.JWTService
	Revoker     model.IJWTRevoker
	ForceLogout model.IForceLogoutStore
	Notifier    model.IEventNotifier
	Now         func() time.Time
}

func (d WSAuthDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

type WSAuthOutput struct {
	UserID uuid.UUID
}

// AuthenticateWebSocket は WS 接続の JWT 検証とセッション登録を行う
func AuthenticateWebSocket(ctx context.Context, d WSAuthDeps, token string) (*WSAuthOutput, error) {
	if token == "" {
		return nil, model.ErrValidation("token is required")
	}
	parsed, err := d.JWT.Parse(token)
	if err != nil {
		return nil, err
	}
	if d.Revoker != nil {
		revoked, err := d.Revoker.IsRevoked(ctx, parsed.JTI)
		if err != nil {
			return nil, model.ErrInternal("failed to check token revocation")
		}
		if revoked {
			return nil, model.ErrUnauthorized()
		}
	}
	if d.ForceLogout != nil && !parsed.IssuedAt.IsZero() {
		before, err := d.ForceLogout.GetForceLogoutBefore(ctx, parsed.UserID)
		if err != nil {
			return nil, model.ErrInternal("failed to check force logout")
		}
		if !before.IsZero() && parsed.IssuedAt.Before(before) {
			return nil, model.ErrUnauthorized()
		}
	}
	user, err := d.Repo.User.FindByID(ctx, parsed.UserID)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, model.ErrUnauthorized()
	}
	return &WSAuthOutput{UserID: parsed.UserID}, nil
}

// NotifyOpponentConnected は対戦中ゲームの相手へ接続状態を通知する
func NotifyOpponentConnected(ctx context.Context, d WSAuthDeps, userID uuid.UUID) {
	active, err := FindActiveGameForUser(ctx, d.Repo, userID)
	if err != nil || active == nil || d.Notifier == nil {
		return
	}
	opponentID, err := active.OpponentID(userID)
	if err != nil {
		return
	}
	_ = d.Notifier.SendToUser(ctx, opponentID, "OPPONENT_STATUS", map[string]any{
		"gameId": active.ID.String(), "playerId": userID.String(), "connected": true,
	})
}
