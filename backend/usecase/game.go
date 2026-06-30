package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// 対戦ゲームのユースケース。
type IGameUsecase interface {
	GetGameState(ctx context.Context, userID, gameID uuid.UUID) (*GameStateOutput, error)
	SyncGameState(ctx context.Context, userID, gameID uuid.UUID) (*GameStateOutput, error)
	SetSecretNumber(ctx context.Context, userID, gameID uuid.UUID, secret string) error
	SubmitGuess(ctx context.Context, userID, gameID uuid.UUID, guess string, isAuto bool) error
	HandleTimeout(ctx context.Context, gameID, playerID uuid.UUID) error
	CancelBySecretTimeout(ctx context.Context, gameID uuid.UUID) error
	RecoverActiveGames(ctx context.Context) error
}

// 秘密数字の hash 化と照合。
type ISecretHasher interface {
	Hash(secret [4]int, gameID uuid.UUID, slot int) (string, error)
	Verify(storedHash string, guess [4]int, gameID uuid.UUID, slot int) ([]model.DigitResult, error)
}

// ゲーム操作の分散ロック。
type IGameLockStore interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type TurnInfo struct {
	Turn      int
	PlayerID  uuid.UUID
	ExpiresAt time.Time
}

// ターンタイマーの Redis 管理。
type ITurnStore interface {
	SetTurn(ctx context.Context, gameID uuid.UUID, turn int, playerID uuid.UUID, startedAt, expiresAt time.Time) error
	GetTurn(ctx context.Context, gameID uuid.UUID) (*TurnInfo, error)
	RemainingSeconds(ctx context.Context, gameID uuid.UUID, now time.Time) (int, error)
	DeleteTurn(ctx context.Context, gameID uuid.UUID) error
}

// 推測数字の乱数生成。
type IGuessNumberGenerator interface {
	GenerateGuessNumber() (string, error)
}

// WebSocket イベント通知。
type IEventNotifier interface {
	SendToUser(ctx context.Context, userID uuid.UUID, eventType string, payload map[string]any) error
}

type GameUseCase struct {
	Games        repository.IGameRepo
	Guesses      repository.IGuessRepo
	Users        repository.IUserRepo
	MatchHistory repository.IMatchHistoryRepo
	ActivityLogs repository.IActivityLogRepo
	Repos        repository.Repos
	Secrets      ISecretHasher
	Locks        IGameLockStore
	Turns        ITurnStore
	Random       IGuessNumberGenerator
	Notifier     IEventNotifier
	TurnDuration time.Duration
	SecretSetup  time.Duration
	GameLockTTL  time.Duration
	Now          func() time.Time
}

func (g *GameUseCase) now() time.Time {
	if g != nil && g.Now != nil {
		return g.Now().UTC()
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

func gameStateWSPayload(s *GameStateOutput) map[string]any {
	guesses := make([]map[string]any, len(s.MyGuesses))
	for i, row := range s.MyGuesses {
		guesses[i] = map[string]any{
			"turn": row.Turn, "guessNumber": row.GuessNumber,
			"digitResults": row.DigitResults, "hitCount": row.HitCount, "isAuto": row.IsAuto,
		}
	}
	return map[string]any{
		"gameId": s.GameID.String(), "status": string(s.Status),
		"currentTurn": s.CurrentTurn, "currentTurnPlayerID": s.CurrentTurnPlayerID,
		"remainingSeconds": s.RemainingSeconds, "myGuesses": guesses,
		"opponentGuessCount": s.OpponentGuessCount,
	}
}

func (g *GameUseCase) GetGameState(ctx context.Context, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	return g.buildGameState(ctx, userID, gameID)
}

func (g *GameUseCase) SyncGameState(ctx context.Context, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	state, err := g.buildGameState(ctx, userID, gameID)
	if err != nil {
		return nil, err
	}
	if g.Notifier != nil {
		_ = g.Notifier.SendToUser(ctx, userID, "GAME_STATE_SYNC", gameStateWSPayload(state))
	}
	return state, nil
}

func (g *GameUseCase) buildGameState(ctx context.Context, userID, gameID uuid.UUID) (*GameStateOutput, error) {
	game, err := g.Games.FindByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrNotFound
	}
	if !game.IsParticipant(userID) {
		return nil, ErrForbidden
	}
	myGuesses, err := g.Guesses.ListByGameAndPlayer(ctx, gameID, userID)
	if err != nil {
		return nil, err
	}
	opponentID, err := gameOpponentID(game, userID)
	if err != nil {
		return nil, err
	}
	oppCount, err := g.Guesses.CountByGameExcludingPlayer(ctx, gameID, opponentID)
	if err != nil {
		return nil, err
	}
	remaining := 0
	if g.Turns != nil && game.Status == model.GameStatusInProgress {
		remaining, err = g.Turns.RemainingSeconds(ctx, gameID, g.now())
		if err != nil {
			return nil, err
		}
	}
	turnPlayer := ""
	if game.CurrentTurnPlayerID != nil {
		turnPlayer = game.CurrentTurnPlayerID.String()
	}
	summaries := make([]GuessSummary, len(myGuesses))
	for i, row := range myGuesses {
		summaries[i] = GuessSummary{
			Turn: row.Turn, GuessNumber: row.GuessNumber,
			DigitResults: DigitResultsToInts(row.DigitResults),
			HitCount:     row.HitCount, IsAuto: row.IsAuto,
		}
	}
	return &GameStateOutput{
		GameID: gameID, Status: game.Status, CurrentTurn: game.CurrentTurn,
		CurrentTurnPlayerID: turnPlayer, RemainingSeconds: remaining,
		MyGuesses: summaries, OpponentGuessCount: int(oppCount),
	}, nil
}

func ensureGamePlayer(ctx context.Context, users repository.IUserRepo, userID uuid.UUID) error {
	user, err := users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil || user.IsDeleted() {
		return ErrUnauthorized
	}
	if user.IsMaster() {
		return ErrForbidden
	}
	return nil
}

func secretLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:secret_lock", gameID, playerID)
}

func guessLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:guess_lock", gameID, playerID)
}

func NewGameUseCase(repos repository.Repos, secrets ISecretHasher, locks IGameLockStore, turns ITurnStore, random IGuessNumberGenerator, notifier IEventNotifier, turnDuration, secretSetup, gameLockTTL time.Duration) *GameUseCase {
	return &GameUseCase{
		Games:        repos.Game,
		Guesses:      repos.Guess,
		Users:        repos.User,
		MatchHistory: repos.MatchHistory,
		ActivityLogs: repos.ActivityLog,
		Repos:        repos,
		Secrets:      secrets,
		Locks:        locks,
		Turns:        turns,
		Random:       random,
		Notifier:     notifier,
		TurnDuration: turnDuration,
		SecretSetup:  secretSetup,
		GameLockTTL:  gameLockTTL,
	}
}
