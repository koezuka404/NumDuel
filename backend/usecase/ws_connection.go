package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// RecordWSConnection は WS 認証成功時に ws_connection_logs へ接続を記録する
func RecordWSConnection(ctx context.Context, d WSAuthDeps, userID uuid.UUID, connectionID string) (uuid.UUID, error) {
	now := d.now()
	id := uuid.New()
	if err := d.Repo.WSConnectionLogs().Create(ctx, &model.WSConnectionLog{
		ID: id, UserID: userID, ConnectionID: connectionID, ConnectedAt: now,
	}); err != nil {
		return uuid.Nil, model.ErrInternal("failed to record ws connection")
	}
	return id, nil
}

// TouchWSActivity は PING / ゲームイベント受信時に last_activity_at を更新する
func TouchWSActivity(ctx context.Context, d WSAuthDeps, userID uuid.UUID) {
	if d.Repo == nil {
		return
	}
	_ = d.Repo.Users().TouchLastActivity(ctx, userID, d.now())
}

// CloseWSConnectionLog は切断時に ws_connection_logs.disconnected_at を更新する
func CloseWSConnectionLog(ctx context.Context, d WSAuthDeps, logID uuid.UUID) {
	if d.Repo == nil || logID == uuid.Nil {
		return
	}
	_ = d.Repo.WSConnectionLogs().UpdateDisconnected(ctx, logID, d.now())
}

// NotifyOpponentDisconnected は対戦中ゲームの相手へ切断を通知する
func NotifyOpponentDisconnected(ctx context.Context, d WSAuthDeps, userID uuid.UUID) {
	active, err := FindActiveGameForUser(ctx, d.Repo, userID)
	if err != nil || active == nil || d.Notifier == nil {
		return
	}
	opponentID, err := active.OpponentID(userID)
	if err != nil {
		return
	}
	_ = d.Notifier.SendToUser(ctx, opponentID, "OPPONENT_STATUS", map[string]any{
		"gameId": active.ID.String(), "playerId": userID.String(), "connected": false,
	})
}
