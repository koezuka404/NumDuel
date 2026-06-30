package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// HandleTimeout はターン期限切れ時に自動予想を 1 回実行する
func HandleTimeout(ctx context.Context, d GameDeps, gameID, playerID uuid.UUID) error {
	now := d.now()
	if d.Locks != nil {
		ok, err := d.Locks.AcquireLock(ctx, guessLockKey(gameID, playerID), d.GameLockTTL)
		if err != nil {
			return model.ErrInternal("failed to acquire guess lock")
		}
		if !ok {
			return nil
		}
	}
	if d.Turns == nil {
		return nil
	}
	turnInfo, err := d.Turns.GetTurn(ctx, gameID)
	if err != nil {
		return model.ErrInternal("failed to read turn deadline")
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
	if err := validateTimeoutGame(ctx, d, gameID, playerID, turnInfo.Turn); err != nil {
		return err
	}
	turnInfo, err = d.Turns.GetTurn(ctx, gameID)
	if err != nil {
		return model.ErrInternal("failed to re-read turn deadline")
	}
	if turnInfo == nil || turnInfo.PlayerID != playerID || turnInfo.ExpiresAt.After(now) {
		return nil
	}
	if d.Random == nil {
		return model.ErrInternal("random number generator is not configured")
	}
	guessNum, err := d.Random.GenerateGuessNumber()
	if err != nil {
		return err
	}
	if err := SubmitGuess(ctx, d, playerID, gameID, guessNum.String(), true); err != nil {
		if isTimeoutRaceError(err) {
			return nil
		}
		return err
	}
	return recordTimeoutActivityLog(ctx, d.Repo, gameID, playerID, now)
}

func validateTimeoutGame(ctx context.Context, d GameDeps, gameID, playerID uuid.UUID, redisTurn int) error {
	game, err := d.Repo.Game.FindByID(ctx, gameID)
	if err != nil {
		return model.ErrInternal("failed to find game")
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
	de, ok := model.IsDomainError(err)
	if !ok {
		return false
	}
	switch de.Code {
	case model.CodeNotYourTurn, model.CodeGameAlreadyFinished, model.CodeGameNotStarted, model.CodeNotFound:
		return true
	default:
		return false
	}
}
