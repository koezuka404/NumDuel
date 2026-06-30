package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// CancelGameBySecretTimeout は秘密数字登録期限切れでゲームを無効終了する
func CancelGameBySecretTimeout(ctx context.Context, d GameDeps, gameID uuid.UUID) error {
	now := d.now()
	var player1, player2 uuid.UUID
	cancelled := false
	if err := repository.WithTx(ctx, d.Repo.DB, func(ctx context.Context) error {
		game, err := d.Repo.Game.FindByIDForUpdate(ctx, gameID)
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
		if err := d.Repo.Game.Update(ctx, game); err != nil {
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
	return recordGameOverActivityLog(ctx, d.Repo, gameID, "secret_setup_timeout", nil, now)
}
