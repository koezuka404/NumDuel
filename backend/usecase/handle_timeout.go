package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func (g *GameUseCase) HandleTimeout(ctx context.Context, gameID, playerID uuid.UUID) error {
	now := g.now()
	if g.Locks != nil {
		ok, err := g.Locks.AcquireLock(ctx, guessLockKey(gameID, playerID), g.GameLockTTL)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	if g.Turns == nil {
		return nil
	}
	turnInfo, err := g.Turns.GetTurn(ctx, gameID)
	if err != nil {
		return err
	}
	if turnInfo == nil {
		return nil
	}
	if turnInfo.PlayerID != playerID {
		return nil
	}
	if turnInfo.ExpiresAt.After(now) {
		return nil
	}
	if err := g.validateTimeoutGame(ctx, gameID, playerID, turnInfo.Turn); err != nil {
		return err
	}
	turnInfo, err = g.Turns.GetTurn(ctx, gameID)
	if err != nil {
		return err
	}
	if turnInfo == nil || turnInfo.PlayerID != playerID || turnInfo.ExpiresAt.After(now) {
		return nil
	}
	if g.Random == nil {
		return errors.New("random number generator is not configured")
	}
	guessNum, err := g.Random.GenerateGuessNumber()
	if err != nil {
		return err
	}
	if err := g.SubmitGuess(ctx, playerID, gameID, guessNum, true); err != nil {
		if isTimeoutRaceError(err) {
			return nil
		}
		return err
	}
	return recordTimeoutActivityLog(ctx, g.Repos, gameID, playerID, now)
}

func (g *GameUseCase) validateTimeoutGame(ctx context.Context, gameID, playerID uuid.UUID, redisTurn int) error {
	game, err := g.Games.FindByID(ctx, gameID)
	if err != nil {
		return err
	}
	if game == nil {
		return nil
	}
	if game.Status != model.GameStatusInProgress {
		return nil
	}
	if game.CurrentTurnPlayerID == nil || *game.CurrentTurnPlayerID != playerID {
		return nil
	}
	if game.CurrentTurn != redisTurn {
		return nil
	}
	return nil
}

func isTimeoutRaceError(err error) bool {
	return errors.Is(err, ErrNotYourTurn) ||
		errors.Is(err, ErrGameAlreadyFinished) ||
		errors.Is(err, ErrGameNotStarted) ||
		errors.Is(err, ErrNotFound)
}
