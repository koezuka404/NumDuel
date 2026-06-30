package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// RecoverActiveGames は起動時に進行中ゲームのターン復元と期限切れ WAITING_SECRET の終了を行う
func RecoverActiveGames(ctx context.Context, d GameDeps) error {
	now := d.now()

	inProgress, err := d.Repo.Games().ListByStatus(ctx, model.GameStatusInProgress)
	if err != nil {
		return model.ErrInternal("failed to list in progress games")
	}
	for _, game := range inProgress {
		if game == nil {
			continue
		}
		if err := recoverInProgressGameTurn(ctx, d, game, now); err != nil {
			return err
		}
	}

	before := now.Add(-d.SecretSetup)
	waitingExpired, err := d.Repo.Games().ListByStatusCreatedBefore(ctx, model.GameStatusWaitingSecret, before)
	if err != nil {
		return model.ErrInternal("failed to list expired waiting games")
	}
	for _, game := range waitingExpired {
		if game == nil {
			continue
		}
		if err := CancelGameBySecretTimeout(ctx, d, game.ID); err != nil {
			return err
		}
	}
	return nil
}

func recoverInProgressGameTurn(ctx context.Context, d GameDeps, game *model.Game, now time.Time) error {
	if d.Turns == nil || game.CurrentTurnPlayerID == nil {
		return nil
	}
	turnInfo, err := d.Turns.GetTurn(ctx, game.ID)
	if err != nil {
		return model.ErrInternal("failed to read turn deadline")
	}
	if turnInfo != nil && turnInfo.ExpiresAt.After(now) {
		return nil
	}

	playerID := *game.CurrentTurnPlayerID
	expiresAt := now.Add(d.TurnDuration)
	if err := d.Turns.SetTurn(ctx, game.ID, game.CurrentTurn, playerID, now, expiresAt); err != nil {
		return model.ErrInternal("failed to restore turn deadline")
	}

	remaining := int(expiresAt.Sub(now).Seconds())
	payload := map[string]any{
		"gameId": game.ID.String(), "currentTurn": game.CurrentTurn,
		"currentTurnPlayerID": playerID.String(),
		"remainingSeconds":    remaining,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		if d.Notifier != nil {
			_ = d.Notifier.SendToUser(ctx, uid, "TURN_CHANGED", payload)
		}
	}
	return recordRecoverActivityLog(ctx, d.Repo, game.ID, now)
}

func recordRecoverActivityLog(ctx context.Context, repo repository.IRepository, gameID uuid.UUID, now time.Time) error {
	detail, err := json.Marshal(map[string]string{"gameId": gameID.String()})
	if err != nil {
		return model.ErrInternal("failed to build activity log")
	}
	if err := repo.ActivityLogs().Create(ctx, &model.ActivityLog{
		ID: uuid.New(), LogType: "recover",
		Detail: detail, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return model.ErrInternal("failed to save activity log")
	}
	return nil
}
