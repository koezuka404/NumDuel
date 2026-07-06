package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// AccessToken解析。
type IAccessTokenParser interface {
	Parse(raw string) (*AccessTokenClaims, error)
}

// WSチケットの発行・解決。ワンタイムでCookieに依存しないWS認証に使う。
type IWSTicketStore interface {
	IssueTicket(ctx context.Context, userID uuid.UUID, ttl time.Duration) (string, error)
	ConsumeTicket(ctx context.Context, ticket string) (uuid.UUID, error)
}

// WebSocket接続認証ユースケース。
type IWSAuthUsecase interface {
	Authenticate(ctx context.Context, token string) (*WSAuthOutput, error)
	AuthenticateByTicket(ctx context.Context, ticket string) (*WSAuthOutput, error)
	IssueTicket(ctx context.Context, userID uuid.UUID) (string, error)
	NotifyOpponentConnected(ctx context.Context, userID uuid.UUID)
	RecordConnection(ctx context.Context, userID uuid.UUID, connectionID string) (uuid.UUID, error)
	TouchActivity(ctx context.Context, userID uuid.UUID)
	CloseConnectionLog(ctx context.Context, logID uuid.UUID)
	NotifyOpponentDisconnected(ctx context.Context, userID uuid.UUID)
}

const wsTicketTTL = 10 * time.Second

type WSAuthUseCase struct {
	Games       repository.IGameRepo
	Users       repository.IUserRepo
	WSLogs      repository.IWSConnectionLogRepo
	Tokens      IAccessTokenParser
	Revoker     IJWTRevoker
	ForceLogout IForceLogoutStore
	Notifier    IEventNotifier
	Tickets     IWSTicketStore
	Now         func() time.Time
}

func NewWSAuthUseCase(repos repository.Repos, tokens IAccessTokenParser, revoker IJWTRevoker, forceLogout IForceLogoutStore, notifier IEventNotifier, tickets IWSTicketStore) *WSAuthUseCase {
	return &WSAuthUseCase{
		Games:       repos.Game,
		Users:       repos.User,
		WSLogs:      repos.WSConnectionLog,
		Tokens:      tokens,
		Revoker:     revoker,
		ForceLogout: forceLogout,
		Notifier:    notifier,
		Tickets:     tickets,
	}
}

func (w *WSAuthUseCase) now() time.Time {
	if w != nil && w.Now != nil {
		return w.Now().UTC()
	}
	return time.Now().UTC()
}

type WSAuthOutput struct {
	UserID uuid.UUID
}

func (w *WSAuthUseCase) Authenticate(ctx context.Context, token string) (*WSAuthOutput, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}
	parsed, err := w.Tokens.Parse(token)
	if err != nil {
		return nil, err
	}
	if w.Revoker != nil {
		revoked, err := w.Revoker.IsRevoked(ctx, parsed.JTI)
		if err != nil {
			return nil, err
		}
		if revoked {
			return nil, ErrUnauthorized
		}
	}
	if w.ForceLogout != nil && !parsed.IssuedAt.IsZero() {
		before, err := w.ForceLogout.GetForceLogoutBefore(ctx, parsed.UserID)
		if err != nil {
			return nil, err
		}
		if !before.IsZero() && parsed.IssuedAt.Before(before) {
			return nil, ErrUnauthorized
		}
	}
	user, err := w.Users.FindByID(ctx, parsed.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted() {
		return nil, ErrUnauthorized
	}
	return &WSAuthOutput{UserID: parsed.UserID}, nil
}

// IssueTicketは同一オリジンHTTP経由(Cookie認証済み)で呼ばれ、WS接続用の使い捨てticketを発行する。
func (w *WSAuthUseCase) IssueTicket(ctx context.Context, userID uuid.UUID) (string, error) {
	if w.Tickets == nil {
		return "", ErrUnauthorized
	}
	return w.Tickets.IssueTicket(ctx, userID, wsTicketTTL)
}

// AuthenticateByTicketはWSハンドシェイク後のAUTHメッセージで渡されたticketを検証する。
// Cookieのcross-site送信制限(SameSite=Lax等)に依存しないための経路。
func (w *WSAuthUseCase) AuthenticateByTicket(ctx context.Context, ticket string) (*WSAuthOutput, error) {
	if ticket == "" || w.Tickets == nil {
		return nil, ErrUnauthorized
	}
	userID, err := w.Tickets.ConsumeTicket(ctx, ticket)
	if err != nil {
		return nil, ErrUnauthorized
	}
	user, err := w.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted() {
		return nil, ErrUnauthorized
	}
	return &WSAuthOutput{UserID: userID}, nil
}

func (w *WSAuthUseCase) findActiveGame(ctx context.Context, userID uuid.UUID) (*model.Game, error) {
	return findActiveGameForUser(ctx, repository.Repos{Game: w.Games, User: w.Users}, userID)
}

func (w *WSAuthUseCase) NotifyOpponentConnected(ctx context.Context, userID uuid.UUID) {
	active, err := w.findActiveGame(ctx, userID)
	if err != nil || active == nil || w.Notifier == nil {
		return
	}
	opponentID, err := gameOpponentID(active, userID)
	if err != nil {
		return
	}
	_ = w.Notifier.SendToUser(ctx, opponentID, "OPPONENT_STATUS", map[string]any{
		"gameId": active.ID.String(), "playerId": userID.String(), "connected": true,
	})
}

func (w *WSAuthUseCase) RecordConnection(ctx context.Context, userID uuid.UUID, connectionID string) (uuid.UUID, error) {
	now := w.now()
	logID := uuid.New()
	if err := w.WSLogs.Create(ctx, &model.WSConnectionLog{
		ID: logID, UserID: userID, ConnectionID: connectionID, ConnectedAt: now,
	}); err != nil {
		return uuid.Nil, err
	}
	return logID, nil
}

func (w *WSAuthUseCase) TouchActivity(ctx context.Context, userID uuid.UUID) {
	user, err := w.Users.FindByID(ctx, userID)
	if err != nil || user == nil || user.IsDeleted() {
		return
	}
	now := w.now()
	user.LastActivityAt = now
	user.UpdatedAt = now
	_ = w.Users.Update(ctx, user)
}

func (w *WSAuthUseCase) CloseConnectionLog(ctx context.Context, logID uuid.UUID) {
	if logID == uuid.Nil {
		return
	}
	_ = w.WSLogs.UpdateDisconnected(ctx, logID, w.now())
}

func (w *WSAuthUseCase) NotifyOpponentDisconnected(ctx context.Context, userID uuid.UUID) {
	active, err := w.findActiveGame(ctx, userID)
	if err != nil || active == nil || w.Notifier == nil {
		return
	}
	opponentID, err := gameOpponentID(active, userID)
	if err != nil {
		return
	}
	_ = w.Notifier.SendToUser(ctx, opponentID, "OPPONENT_STATUS", map[string]any{
		"gameId": active.ID.String(), "playerId": userID.String(), "connected": false,
	})
}
