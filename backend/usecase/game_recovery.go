package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func (g *GameUseCase) CancelBySecretTimeout(ctx context.Context, gameID uuid.UUID) error {
	now := g.now()
	var player1, player2 uuid.UUID
	cancelled := false
	if err := repository.WithTx(ctx, g.Repos.DB, func(ctx context.Context) error {
		game, err := g.Games.FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return err
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
		deadline := game.CreatedAt.Add(g.SecretSetup)
		if !now.After(deadline) {
			return nil
		}
		cancelGameBySecretTimeout(game, now)
		if err := g.Games.Update(ctx, game); err != nil {
			return err
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
	if g.Turns != nil {
		_ = g.Turns.DeleteTurn(ctx, gameID)
	}
	over := map[string]any{"gameId": gameID.String(), "reason": "secret_setup_timeout"}
	for _, uid := range []uuid.UUID{player1, player2} {
		if g.Notifier != nil {
			_ = g.Notifier.SendToUser(ctx, uid, "GAME_OVER", over)
		}
	}
	return recordGameOverActivityLog(ctx, g.Repos, gameID, "secret_setup_timeout", nil, now)
}

func (g *GameUseCase) RecoverActiveGames(ctx context.Context) error {
	now := g.now()
	inProgress, err := g.Games.ListByStatus(ctx, model.GameStatusInProgress)
	if err != nil {
		return err
	}
	for _, game := range inProgress {
		if game == nil {
			continue
		}
		if err := g.recoverInProgressGameTurn(ctx, game, now); err != nil {
			return err
		}
	}
	before := now.Add(-g.SecretSetup)
	waitingExpired, err := g.Games.ListByStatusCreatedBefore(ctx, model.GameStatusWaitingSecret, before)
	if err != nil {
		return err
	}
	for _, game := range waitingExpired {
		if game == nil {
			continue
		}
		if err := g.CancelBySecretTimeout(ctx, game.ID); err != nil {
			return err
		}
	}
	return nil
}

func (g *GameUseCase) recoverInProgressGameTurn(ctx context.Context, game *model.Game, now time.Time) error {
	if g.Turns == nil || game.CurrentTurnPlayerID == nil {
		return nil
	}
	turnInfo, err := g.Turns.GetTurn(ctx, game.ID)
	if err != nil {
		return err
	}
	if turnInfo != nil && turnInfo.ExpiresAt.After(now) {
		return nil
	}
	playerID := *game.CurrentTurnPlayerID
	expiresAt := now.Add(g.TurnDuration)
	if err := g.Turns.SetTurn(ctx, game.ID, game.CurrentTurn, playerID, now, expiresAt); err != nil {
		return err
	}
	remaining := int(expiresAt.Sub(now).Seconds())
	payload := map[string]any{
		"gameId": game.ID.String(), "currentTurn": game.CurrentTurn,
		"currentTurnPlayerID": playerID.String(),
		"remainingSeconds":    remaining,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		if g.Notifier != nil {
			_ = g.Notifier.SendToUser(ctx, uid, "TURN_CHANGED", payload)
		}
	}
	return recordRecoverActivityLog(ctx, g.Repos, game.ID, now)
}
