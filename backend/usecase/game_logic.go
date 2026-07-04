package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func gamePlayerSlot(game *model.Game, userID uuid.UUID) (int, error) {
	switch userID {
	case game.Player1ID:
		return 1, nil
	case game.Player2ID:
		return 2, nil
	default:
		return 0, ErrForbidden
	}
}

func setGameSecretHash(game *model.Game, userID uuid.UUID, hash string) error {
	switch userID {
	case game.Player1ID:
		if game.Player1Secret != "" {
			return ErrBadRequest
		}
		game.Player1Secret = hash
		return nil
	case game.Player2ID:
		if game.Player2Secret != "" {
			return ErrBadRequest
		}
		game.Player2Secret = hash
		return nil
	default:
		return ErrForbidden
	}
}

func startGame(game *model.Game, now time.Time) error {
	if game.Status != model.GameStatusWaitingSecret || !game.BothSecretsSet() {
		return ErrBadRequest
	}
	game.Status = model.GameStatusInProgress
	game.CurrentTurn = 1
	game.CurrentTurnPlayerID = &game.Player1ID
	game.StartedAt = &now
	game.UpdatedAt = now
	return nil
}

func gameOpponentID(game *model.Game, userID uuid.UUID) (uuid.UUID, error) {
	switch userID {
	case game.Player1ID:
		return game.Player2ID, nil
	case game.Player2ID:
		return game.Player1ID, nil
	default:
		return uuid.Nil, ErrForbidden
	}
}

func gameOpponentSecretHash(game *model.Game, userID uuid.UUID) (string, int, error) {
	switch userID {
	case game.Player1ID:
		return game.Player2Secret, 2, nil
	case game.Player2ID:
		return game.Player1Secret, 1, nil
	default:
		return "", 0, ErrForbidden
	}
}

func addGameGuess(game *model.Game, userID uuid.UUID, guessDigits [4]int, results [4]model.DigitResult, isAuto bool, now time.Time) (model.Guess, error) {
	if game.Status != model.GameStatusInProgress {
		return model.Guess{}, ErrGameNotStarted
	}
	if !game.CanGuess(userID) {
		return model.Guess{}, ErrNotYourTurn
	}
	guessNumber := fmt.Sprintf("%d%d%d%d", guessDigits[0], guessDigits[1], guessDigits[2], guessDigits[3])
	g := model.Guess{
		ID:           uuid.New(),
		GameID:       game.ID,
		PlayerID:     userID,
		Turn:         game.CurrentTurn,
		GuessNumber:  guessNumber,
		DigitResults: results[:],
		HitCount:     HitCount(results),
		IsAuto:       isAuto,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	nextPlayer, err := gameOpponentID(game, userID)
	if err != nil {
		return model.Guess{}, err
	}
	game.CurrentTurn++
	game.CurrentTurnPlayerID = &nextPlayer
	game.UpdatedAt = now
	return g, nil
}

func finishGame(game *model.Game, winnerID uuid.UUID, now time.Time) error {
	if game.Status != model.GameStatusInProgress {
		return ErrGameAlreadyFinished
	}
	if winnerID != game.Player1ID && winnerID != game.Player2ID {
		return errors.New("勝者が参加者ではありません")
	}
	game.Status = model.GameStatusFinished
	game.WinnerID = &winnerID
	game.FinishedAt = &now
	game.UpdatedAt = now
	return nil
}

func cancelGameBySecretTimeout(game *model.Game, now time.Time) error {
	if game.Status != model.GameStatusWaitingSecret {
		return ErrBadRequest
	}
	game.Status = model.GameStatusFinished
	game.FinishedAt = &now
	game.UpdatedAt = now
	return nil
}
