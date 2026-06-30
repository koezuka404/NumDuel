package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// CancelGameBySecretTimeout は秘密数字登録期限切れでゲームを無効終了する
func CancelGameBySecretTimeout(ctx context.Context, d GameDeps, gameID uuid.UUID) error {
	now := d.now()
	var player1, player2 uuid.UUID
	cancelled := false
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		game, err := tx.Games().FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return model.ErrInternal("failed to find game")
		}
		if game == nil {
			return nil
		}
		if game.Status != model.GameStatusWaitingSecret {
			return nil
		}
		if game.StartedAt != nil || game.FinishedAt != nil || game.WinnerID != nil {
			return nil
		}
		deadline := game.CreatedAt.Add(d.SecretSetup)
		if !now.After(deadline) {
			return nil
		}
		if err := game.CancelBySecretTimeout(now); err != nil {
			return err
		}
		if err := tx.Games().Update(ctx, game); err != nil {
			return model.ErrInternal("failed to cancel game")
		}
		player1, player2 = game.Player1ID, game.Player2ID
		cancelled = true
		return nil
	}); err != nil {
		return err
	}
	if !cancelled {
		return nil
	}
	if d.Turns != nil {
		_ = d.Turns.DeleteTurn(ctx, gameID)
	}
	over := map[string]any{
		"gameId": gameID.String(), "reason": "secret_setup_timeout",
	}
	for _, uid := range []uuid.UUID{player1, player2} {
		if d.Notifier != nil {
			_ = d.Notifier.SendToUser(ctx, uid, "GAME_OVER", over)
		}
	}
	return recordGameOverActivityLog(ctx, d.Repo, gameID, "secret_setup_timeout", now)
}

func recordGameOverActivityLog(ctx context.Context, repo repository.IRepository, gameID uuid.UUID, reason string, now time.Time) error {
	detail, err := json.Marshal(map[string]string{
		"gameId": gameID.String(), "reason": reason,
	})
	if err != nil {
		return model.ErrInternal("failed to build activity log")
	}
	if err := repo.ActivityLogs().Create(ctx, &model.ActivityLog{
		ID: uuid.New(), LogType: "game_over",
		Detail: detail, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return model.ErrInternal("failed to save activity log")
	}
	return nil
}
