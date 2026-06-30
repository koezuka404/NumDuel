package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// GameDeps はゲーム系 UseCase の依存関係
type GameDeps struct {
	Repo         repository.IRepository
	Tx           repository.ITxManager
	Secrets      model.ISecretHasher
	Locks        model.IGameLockStore
	Turns        model.ITurnStore
	Random       model.IGuessNumberGenerator
	Notifier     model.IEventNotifier
	TurnDuration time.Duration
	SecretSetup  time.Duration
	GameLockTTL  time.Duration
	Now          func() time.Time
}

func (d GameDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

type GuessSummary struct {
	Turn         int
	GuessNumber  string
	DigitResults []int
	HitCount     int
	IsAuto       bool
}

type GameStateOutput struct {
	GameID              uuid.UUID
	Status              model.GameStatus
	CurrentTurn         int
	CurrentTurnPlayerID string
	RemainingSeconds    int
	MyGuesses           []GuessSummary
	OpponentGuessCount  int
}

func GameStateToMap(s *GameStateOutput) map[string]any {
	turnPlayer := s.CurrentTurnPlayerID
	guesses := make([]map[string]any, len(s.MyGuesses))
	for i, g := range s.MyGuesses {
		guesses[i] = map[string]any{
			"turn": g.Turn, "guessNumber": g.GuessNumber,
			"digitResults": g.DigitResults, "hitCount": g.HitCount, "isAuto": g.IsAuto,
		}
	}
	return map[string]any{
		"gameId": s.GameID.String(), "status": string(s.Status),
		"currentTurn": s.CurrentTurn, "currentTurnPlayerID": turnPlayer,
		"remainingSeconds": s.RemainingSeconds, "myGuesses": guesses,
		"opponentGuessCount": s.OpponentGuessCount,
	}
}

func GetGameState(ctx context.Context, d GameDeps, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	return buildGameState(ctx, d, userID, gameID)
}

func SyncGameState(ctx context.Context, d GameDeps, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	state, err := buildGameState(ctx, d, userID, gameID)
	if err != nil {
		return nil, err
	}
	if d.Notifier != nil {
		_ = d.Notifier.SendToUser(ctx, userID, "GAME_STATE_SYNC", GameStateToMap(state))
	}
	return state, nil
}

func buildGameState(ctx context.Context, d GameDeps, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	game, err := d.Repo.Games().FindByID(ctx, gameID)
	if err != nil {
		return nil, model.ErrInternal("failed to find game")
	}
	if game == nil {
		return nil, model.ErrNotFound("game not found")
	}
	if !game.IsParticipant(userID) {
		return nil, model.ErrForbidden("not a participant")
	}
	myGuesses, err := d.Repo.Guesses().ListByGameAndPlayer(ctx, gameID, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to load guesses")
	}
	opponentID, err := game.OpponentID(userID)
	if err != nil {
		return nil, err
	}
	oppCount, err := d.Repo.Guesses().CountByGameExcludingPlayer(ctx, gameID, opponentID)
	if err != nil {
		return nil, model.ErrInternal("failed to count opponent guesses")
	}
	remaining := 0
	if d.Turns != nil && game.Status == model.GameStatusInProgress {
		remaining, err = d.Turns.RemainingSeconds(ctx, gameID, d.now())
		if err != nil {
			return nil, model.ErrInternal("failed to read turn deadline")
		}
	}
	turnPlayer := ""
	if game.CurrentTurnPlayerID != nil {
		turnPlayer = game.CurrentTurnPlayerID.String()
	}
	summaries := make([]GuessSummary, len(myGuesses))
	for i, g := range myGuesses {
		summaries[i] = GuessSummary{
			Turn: g.Turn, GuessNumber: g.GuessNumber,
			DigitResults: model.DigitResultsToInts(g.DigitResults),
			HitCount:     g.HitCount, IsAuto: g.IsAuto,
		}
	}
	return &GameStateOutput{
		GameID: gameID, Status: game.Status, CurrentTurn: game.CurrentTurn,
		CurrentTurnPlayerID: turnPlayer, RemainingSeconds: remaining,
		MyGuesses: summaries, OpponentGuessCount: int(oppCount),
	}, nil
}

func SetSecretNumber(ctx context.Context, d GameDeps, userID, gameID uuid.UUID, secret string) error {
	if d.Locks != nil {
		ok, err := d.Locks.AcquireLock(ctx, secretLockKey(gameID, userID), d.GameLockTTL)
		if err != nil {
			return model.ErrInternal("failed to acquire secret lock")
		}
		if !ok {
			return model.ErrRateLimitExceeded()
		}
	}
	secretNum, err := model.NewSecretNumberFromString(secret)
	if err != nil {
		return err
	}
	if err := ensureGamePlayer(ctx, d.Repo, userID); err != nil {
		return err
	}
	var started bool
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		game, err := tx.Games().FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return model.ErrInternal("failed to find game")
		}
		if game == nil {
			return model.ErrNotFound("game not found")
		}
		if !game.IsParticipant(userID) {
			return model.ErrForbidden("not a participant")
		}
		switch game.Status {
		case model.GameStatusInProgress:
			return model.ErrGameAlreadyStarted()
		case model.GameStatusFinished:
			return model.ErrGameAlreadyFinished()
		case model.GameStatusWaitingSecret:
		default:
			return model.ErrValidation("invalid game status")
		}
		slot, err := game.PlayerSlot(userID)
		if err != nil {
			return err
		}
		hash, err := d.Secrets.Hash(secretNum.Digits(), gameID, slot)
		if err != nil {
			return model.ErrInternal("failed to hash secret")
		}
		if err := game.SetSecretHash(userID, hash); err != nil {
			return err
		}
		game.UpdatedAt = d.now()
		if err := tx.Games().Update(ctx, game); err != nil {
			return model.ErrInternal("failed to save secret")
		}
		if game.BothSecretsSet() {
			if err := startGameInTx(ctx, tx, game, d.now()); err != nil {
				return err
			}
			started = true
		}
		return nil
	}); err != nil {
		return err
	}
	if started {
		return notifyGameStart(ctx, d, gameID)
	}
	return nil
}

func startGameInTx(ctx context.Context, tx repository.ITxRepos, game *model.Game, now time.Time) error {
	if err := game.Start(now); err != nil {
		return err
	}
	return tx.Games().Update(ctx, game)
}

func notifyGameStart(ctx context.Context, d GameDeps, gameID uuid.UUID) error {
	game, err := d.Repo.Games().FindByID(ctx, gameID)
	if err != nil || game == nil {
		return model.ErrInternal("failed to load started game")
	}
	now := d.now()
	expires := now.Add(d.TurnDuration)
	if d.Turns != nil && game.CurrentTurnPlayerID != nil {
		if err := d.Turns.SetTurn(ctx, gameID, game.CurrentTurn, *game.CurrentTurnPlayerID, now, expires); err != nil {
			return model.ErrInternal("failed to set turn deadline")
		}
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		state, err := buildGameState(ctx, d, uid, gameID)
		if err != nil {
			return err
		}
		if d.Notifier != nil {
			_ = d.Notifier.SendToUser(ctx, uid, "GAME_STATE_SYNC", GameStateToMap(state))
		}
	}
	return notifyTurnChanged(ctx, d, game)
}

func notifyTurnChanged(ctx context.Context, d GameDeps, game *model.Game) error {
	if d.Notifier == nil || game.CurrentTurnPlayerID == nil {
		return nil
	}
	remaining := 0
	if d.Turns != nil {
		var err error
		remaining, err = d.Turns.RemainingSeconds(ctx, game.ID, d.now())
		if err != nil {
			return model.ErrInternal("failed to read turn deadline")
		}
	}
	payload := map[string]any{
		"gameId": game.ID.String(), "currentTurn": game.CurrentTurn,
		"currentTurnPlayerID": game.CurrentTurnPlayerID.String(),
		"remainingSeconds":    remaining,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		_ = d.Notifier.SendToUser(ctx, uid, "TURN_CHANGED", payload)
	}
	return nil
}

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

func SubmitGuess(ctx context.Context, d GameDeps, userID, gameID uuid.UUID, guess string, isAuto bool) error {
	if d.Locks != nil && !isAuto {
		ok, err := d.Locks.AcquireLock(ctx, guessLockKey(gameID, userID), d.GameLockTTL)
		if err != nil {
			return model.ErrInternal("failed to acquire guess lock")
		}
		if !ok {
			return model.ErrRateLimitExceeded()
		}
	}
	guessNum, err := model.NewGuessNumberFromString(guess)
	if err != nil {
		return err
	}
	if err := ensureGamePlayer(ctx, d.Repo, userID); err != nil {
		return err
	}
	var result guessResultOutput
	var finished bool
	var winnerID uuid.UUID
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		game, err := tx.Games().FindByIDForUpdate(ctx, gameID)
		if err != nil {
			return model.ErrInternal("failed to find game")
		}
		if game == nil {
			return model.ErrNotFound("game not found")
		}
		if !game.IsParticipant(userID) {
			return model.ErrForbidden("not a participant")
		}
		opponentHash, opponentSlot, err := game.OpponentSecretHash(userID)
		if err != nil {
			return err
		}
		if opponentHash == "" {
			return model.ErrValidation("opponent secret not registered")
		}
		results, err := d.Secrets.Verify(opponentHash, guessNum, gameID, opponentSlot)
		if err != nil {
			return err
		}
		now := d.now()
		g, err := game.AddGuess(userID, guessNum, results, isAuto, now)
		if err != nil {
			return err
		}
		if err := tx.Guesses().Create(ctx, &g); err != nil {
			return model.ErrInternal("failed to save guess")
		}
		isWin := model.IsWin(results)
		result = guessResultOutput{
			GuessID: g.ID, PlayerID: userID, Turn: g.Turn,
			DigitResults: model.DigitResultsToInts(g.DigitResults),
			HitCount:     g.HitCount, IsWin: isWin, IsAuto: isAuto,
		}
		if isWin {
			if err := finishGameService(ctx, tx, game, userID, now); err != nil {
				return err
			}
			finished = true
			winnerID = userID
		} else if err := tx.Games().Update(ctx, game); err != nil {
			return model.ErrInternal("failed to update game turn")
		}
		if game.CurrentTurnPlayerID != nil {
			result.NextTurnPlayerID = game.CurrentTurnPlayerID.String()
		}
		return nil
	}); err != nil {
		return err
	}
	now := d.now()
	if err := recordGuessActivityLog(ctx, d.Repo, gameID, userID, result.Turn, result.HitCount, result.IsWin, isAuto, now); err != nil {
		return err
	}
	return notifyGuessResult(ctx, d, gameID, result, finished, winnerID)
}

func finishGameService(ctx context.Context, tx repository.ITxRepos, game *model.Game, winnerID uuid.UUID, now time.Time) error {
	if winnerID == uuid.Nil {
		return model.ErrInternal("winner is required")
	}
	loserID, err := game.OpponentID(winnerID)
	if err != nil {
		return err
	}
	winner, err := tx.Users().FindByID(ctx, winnerID)
	if err != nil || winner == nil {
		return model.ErrInternal("failed to find winner")
	}
	loser, err := tx.Users().FindByID(ctx, loserID)
	if err != nil || loser == nil {
		return model.ErrInternal("failed to find loser")
	}
	if err := game.Finish(winnerID, now); err != nil {
		return err
	}
	if err := tx.Games().Update(ctx, game); err != nil {
		return model.ErrInternal("failed to finish game")
	}
	history := model.NewMatchHistory(game.ID, winnerID, loserID, winner.Username, loser.Username, now)
	if err := tx.MatchHistories().Create(ctx, &history); err != nil {
		return model.ErrInternal("failed to create match history")
	}
	if err := incrementUserWinCount(ctx, tx, winnerID, now); err != nil {
		return model.ErrInternal("failed to increment win count")
	}
	return nil
}

func notifyGuessResult(ctx context.Context, d GameDeps, gameID uuid.UUID, result guessResultOutput, finished bool, winnerID uuid.UUID) error {
	game, err := d.Repo.Games().FindByID(ctx, gameID)
	if err != nil || game == nil {
		return model.ErrInternal("failed to load game")
	}
	guessPayload := map[string]any{
		"gameId": gameID.String(), "guessId": result.GuessID.String(),
		"playerId": result.PlayerID.String(), "digitResults": result.DigitResults,
		"hitCount": result.HitCount, "isWin": result.IsWin, "isAuto": result.IsAuto,
		"nextTurnPlayerID": result.NextTurnPlayerID,
	}
	for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
		if d.Notifier != nil {
			_ = d.Notifier.SendToUser(ctx, uid, "GUESS_RESULT", guessPayload)
		}
	}
	if finished {
		if d.Turns != nil {
			_ = d.Turns.DeleteTurn(ctx, gameID)
		}
		if err := recordGameOverActivityLog(ctx, d.Repo, gameID, "guess_win", &winnerID, d.now()); err != nil {
			return err
		}
		over := map[string]any{
			"gameId": gameID.String(), "reason": "guess_win", "winnerId": winnerID.String(),
		}
		for _, uid := range []uuid.UUID{game.Player1ID, game.Player2ID} {
			if d.Notifier != nil {
				_ = d.Notifier.SendToUser(ctx, uid, "GAME_OVER", over)
			}
		}
		return nil
	}
	if d.Turns != nil && game.CurrentTurnPlayerID != nil {
		now := d.now()
		if err := d.Turns.SetTurn(ctx, gameID, game.CurrentTurn, *game.CurrentTurnPlayerID, now, now.Add(d.TurnDuration)); err != nil {
			return model.ErrInternal("failed to set next turn")
		}
	}
	return notifyTurnChanged(ctx, d, game)
}

func ensureGamePlayer(ctx context.Context, repo repository.IRepository, userID uuid.UUID) error {
	user, err := repo.Users().FindByID(ctx, userID)
	if err != nil {
		return model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return model.ErrUnauthorized()
	}
	if user.IsMaster() {
		return model.ErrForbidden("master cannot play")
	}
	return nil
}

func secretLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:secret_lock", gameID, playerID)
}

func guessLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:guess_lock", gameID, playerID)
}
