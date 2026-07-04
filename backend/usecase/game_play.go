package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type guessResultOutput struct {
	GuessID          uuid.UUID
	PlayerID         uuid.UUID
	Turn             int
	DigitResults     []int
	HitCount         int
	IsWin            bool
	IsAuto           bool
	NextTurnPlayerID string
}

func (g *GameUseCase) SetSecretNumber(ctx context.Context, userID, gameID uuid.UUID, secret string) error {
	if g.Locks != nil {
		ok, err := g.Locks.AcquireLock(ctx, secretLockKey(gameID, userID), g.GameLockTTL)
		if err != nil {
			return err
		}
		if !ok {
			return ErrRateLimitExceeded
		}
	}
	secretDigits, err := ValidateFourDigits(secret)
	if err != nil {
		return err
	}
	if err := ensureGamePlayer(ctx, g.Users, userID); err != nil {
		return err
	}
	var started bool
	if err := repository.WithTx(ctx, g.Repos.DB, func(ctx context.Context) error {
		game, err := g.Games.FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return err
		}
		if game == nil {
			return ErrNotFound
		}
		if !game.IsParticipant(userID) {
			return ErrForbidden
		}
		switch game.Status {
		case model.GameStatusInProgress:
			return ErrGameAlreadyStarted
		case model.GameStatusFinished:
			return ErrGameAlreadyFinished
		case model.GameStatusWaitingSecret:
		default:
			return ErrBadRequest
		}
		slot := 1
		if userID == game.Player2ID {
			slot = 2
		}
		hash, err := g.Secrets.Hash(secretDigits, gameID, slot)
		if err != nil {
			return err
		}
		if err := setGameSecretHash(game, userID, hash); err != nil {
			return err
		}
		game.UpdatedAt = g.now()
		if err := g.Games.Update(ctx, game); err != nil {
			return err
		}
		if game.BothSecretsSet() {
			if err := startGameInTx(ctx, g.Games, game, g.now()); err != nil {
				return err
			}
			started = true
		}
		return nil
	}); err != nil {
		return err
	}
	if started {
		return g.notifyGameStart(ctx, gameID)
	}
	return nil
}

func startGameInTx(ctx context.Context, games repository.IGameRepo, game *model.Game, now time.Time) error {
	if err := startGame(game, now); err != nil {
		return err
	}
	return games.Update(ctx, game)
}

func (g *GameUseCase) SubmitGuess(ctx context.Context, userID, gameID uuid.UUID, guess string, isAuto bool) error {
	if g.Locks != nil && !isAuto {
		ok, err := g.Locks.AcquireLock(ctx, guessLockKey(gameID, userID), g.GameLockTTL)
		if err != nil {
			return err
		}
		if !ok {
			return ErrRateLimitExceeded
		}
	}
	guessDigits, err := ValidateFourDigits(guess)
	if err != nil {
		return err
	}
	if err := ensureGamePlayer(ctx, g.Users, userID); err != nil {
		return err
	}
	var result guessResultOutput
	var finished bool
	var winnerID uuid.UUID
	if err := repository.WithTx(ctx, g.Repos.DB, func(ctx context.Context) error {
		game, err := g.Games.FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return err
		}
		if game == nil {
			return ErrNotFound
		}
		if !game.IsParticipant(userID) {
			return ErrForbidden
		}
		opponentHash := game.Player2Secret
		opponentSlot := 2
		if userID == game.Player2ID {
			opponentHash = game.Player1Secret
			opponentSlot = 1
		}
		if opponentHash == "" {
			return ErrBadRequest
		}
		resultsSlice, err := g.Secrets.Verify(opponentHash, guessDigits, gameID, opponentSlot)
		if err != nil {
			return err
		}
		var results [4]model.DigitResult
		for i := 0; i < 4 && i < len(resultsSlice); i++ {
			results[i] = resultsSlice[i]
		}
		now := g.now()
		row, err := addGameGuess(game, userID, guessDigits, results, isAuto, now)
		if err != nil {
			return err
		}
		if err := g.Guesses.Create(ctx, &row); err != nil {
			return err
		}
		isWin := IsWin(results)
		result = guessResultOutput{
			GuessID: row.ID, PlayerID: userID, Turn: row.Turn,
			DigitResults: DigitResultsToInts(row.DigitResults),
			HitCount:     row.HitCount, IsWin: isWin, IsAuto: isAuto,
		}
		if isWin {
			if err := finishGameService(ctx, g, game, userID, now); err != nil {
				return err
			}
			finished = true
			winnerID = userID
		} else if err := g.Games.Update(ctx, game); err != nil {
			return err
		}
		if game.CurrentTurnPlayerID != nil {
			result.NextTurnPlayerID = game.CurrentTurnPlayerID.String()
		}
		return nil
	}); err != nil {
		return err
	}
	now := g.now()
	if err := recordGuessActivityLog(ctx, g.Repos, gameID, userID, result.Turn, result.HitCount, result.IsWin, isAuto, now); err != nil {
		return err
	}
	return g.notifyGuessResult(ctx, gameID, result, finished, winnerID)
}

func finishGameService(ctx context.Context, g *GameUseCase, game *model.Game, winnerID uuid.UUID, now time.Time) error {
	if winnerID == uuid.Nil {
		return ErrBadRequest
	}
	loserID, err := gameOpponentID(game, winnerID)
	if err != nil {
		return err
	}
	winner, err := g.Users.FindByID(ctx, winnerID)
	if err != nil || winner == nil {
		return mapRepoNotFound(err)
	}
	loser, err := g.Users.FindByID(ctx, loserID)
	if err != nil || loser == nil {
		return mapRepoNotFound(err)
	}
	if err := finishGame(game, winnerID, now); err != nil {
		return err
	}
	if err := g.Games.Update(ctx, game); err != nil {
		return err
	}
	history := model.MatchHistory{
		ID: uuid.New(), GameID: game.ID, WinnerID: winnerID, LoserID: loserID,
		WinnerUsername: winner.Username, LoserUsername: loser.Username,
		FinishedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := g.MatchHistory.Create(ctx, &history); err != nil {
		return err
	}
	return incrementUserWinCount(ctx, g.Repos, winnerID, now)
}

func incrementUserWinCount(ctx context.Context, repos repository.Repos, userID uuid.UUID, now time.Time) error {
	user, err := repos.User.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotFound
	}
	user.WinCount++
	user.UpdatedAt = now
	return repos.User.Update(ctx, user)
}

func (g *GameUseCase) notifyGameStart(ctx context.Context, gameID uuid.UUID) error {
	game, err := g.Games.FindByID(ctx, gameID)
	if err != nil || game == nil {
		return err
	}
	now := g.now()
	expires := now.Add(g.TurnDuration)
	if g.Turns != nil && game.CurrentTurnPlayerID != nil {
		if err := g.Turns.SetTurn(ctx, gameID, game.CurrentTurn, *game.CurrentTurnPlayerID, now, expires); err != nil {
			return err
		}
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		state, err := g.buildGameState(ctx, uid, gameID)
		if err != nil {
			return err
		}
		if g.Notifier != nil {
			_ = g.Notifier.SendToUser(ctx, uid, "GAME_STATE_SYNC", gameStateWSPayload(state))
		}
	}
	return g.notifyTurnChanged(ctx, game)
}

func (g *GameUseCase) notifyTurnChanged(ctx context.Context, game *model.Game) error {
	if g.Notifier == nil || game.CurrentTurnPlayerID == nil {
		return nil
	}
	remaining := 0
	if g.Turns != nil {
		var err error
		remaining, err = g.Turns.RemainingSeconds(ctx, game.ID, g.now())
		if err != nil {
			return err
		}
	}
	payload := map[string]any{
		"gameId": game.ID.String(), "currentTurn": game.CurrentTurn,
		"currentTurnPlayerID": game.CurrentTurnPlayerID.String(),
		"remainingSeconds":    remaining,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		_ = g.Notifier.SendToUser(ctx, uid, "TURN_CHANGED", payload)
	}
	return nil
}

func (g *GameUseCase) notifyGuessResult(ctx context.Context, gameID uuid.UUID, result guessResultOutput, finished bool, winnerID uuid.UUID) error {
	game, err := g.Games.FindByID(ctx, gameID)
	if err != nil || game == nil {
		return err
	}
	guessPayload := map[string]any{
		"gameId": gameID.String(), "guessId": result.GuessID.String(),
		"playerId": result.PlayerID.String(), "digitResults": result.DigitResults,
		"hitCount": result.HitCount, "isWin": result.IsWin, "isAuto": result.IsAuto,
		"nextTurnPlayerID": result.NextTurnPlayerID,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		if g.Notifier != nil {
			_ = g.Notifier.SendToUser(ctx, uid, "GUESS_RESULT", guessPayload)
		}
	}
	if finished {
		if g.Turns != nil {
			_ = g.Turns.DeleteTurn(ctx, gameID)
		}
		if err := recordGameOverActivityLog(ctx, g.Repos, gameID, "guess_win", &winnerID, g.now()); err != nil {
			return err
		}
		over := map[string]any{
			"gameId": gameID.String(), "reason": "guess_win", "winnerId": winnerID.String(),
		}
		for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
			if g.Notifier != nil {
				_ = g.Notifier.SendToUser(ctx, uid, "GAME_OVER", over)
			}
		}
		return nil
	}
	if g.Turns != nil && game.CurrentTurnPlayerID != nil {
		now := g.now()
		if err := g.Turns.SetTurn(ctx, gameID, game.CurrentTurn, *game.CurrentTurnPlayerID, now, now.Add(g.TurnDuration)); err != nil {
			return err
		}
	}
	return g.notifyTurnChanged(ctx, game)
}
