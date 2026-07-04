package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func TestSetSecretNumberValidateFourDigitsError(t *testing.T) {
	err := (&GameUseCase{}).SetSecretNumber(context.Background(), uuid.New(), uuid.New(), "12ab")
	if !errors.Is(err, model.ErrBadDigit) {
		t.Fatalf("invalid digits: %v", err)
	}
}

func TestFinishGameServiceNilWinner(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CreatedAt: now, UpdatedAt: now,
	}
	g := &GameUseCase{}
	err := finishGameService(context.Background(), g, game, uuid.Nil, now)
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("nil winner: %v", err)
	}
}

func TestStartGameInTxSuccess(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusWaitingSecret,
		Player1ID: p1, Player2ID: p2,
		Player1Secret: "h1", Player2Secret: "h2",
		CurrentTurn: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := startGameInTx(context.Background(), repos.Game, game, now); err != nil {
		t.Fatalf("start game in tx: %v", err)
	}
	if game.Status != model.GameStatusInProgress {
		t.Fatalf("status: %s", game.Status)
	}
}

func TestStartGameInTxInvalidState(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	game := &model.Game{ID: uuid.New(), Status: model.GameStatusWaitingSecret, CreatedAt: now, UpdatedAt: now}
	err := startGameInTx(context.Background(), repos.Game, game, now)
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("invalid start: %v", err)
	}
}

func TestIncrementUserWinCountNotFound(t *testing.T) {
	repos := openTestRepos(t)
	err := incrementUserWinCount(context.Background(), repos, uuid.New(), time.Now().UTC())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing user: %v", err)
	}
}

func TestFinishGameServiceSuccess(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p1, Username: "alice", Email: "a@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p2, Username: "bob", Email: "b@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create p2: %v", err)
	}
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Game.Create(context.Background(), game); err != nil {
		t.Fatalf("create game: %v", err)
	}
	g := &GameUseCase{Games: repos.Game, Users: repos.User, MatchHistory: repos.MatchHistory, Repos: repos}
	if err := finishGameService(context.Background(), g, game, p1, now); err != nil {
		t.Fatalf("finish: %v", err)
	}
	winner, _ := repos.User.FindByID(context.Background(), p1)
	if winner.WinCount != 1 {
		t.Fatalf("win count: %d", winner.WinCount)
	}
}

func TestSetGameSecretHashPlayer2Duplicate(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{Player1ID: p1, Player2ID: p2, Player2Secret: "h2", CreatedAt: now, UpdatedAt: now}
	if err := setGameSecretHash(game, p2, "h3"); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("duplicate p2 secret: %v", err)
	}
}

func TestFinishGameServiceWinnerNilFind(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p1, Username: "alice", Email: "a@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p2, Username: "bob", Email: "b@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create p2: %v", err)
	}
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Game.Create(context.Background(), game); err != nil {
		t.Fatalf("create game: %v", err)
	}
	g := &GameUseCase{
		Games: repos.Game, Users: nilWinnerUserFindRepo{IUserRepo: repos.User, nilID: p1},
		MatchHistory: repos.MatchHistory, Repos: repos,
	}
	if err := finishGameService(context.Background(), g, game, p1, now); err != nil {
		t.Fatalf("winner nil find: %v", err)
	}
}

type nilWinnerUserFindRepo struct {
	repository.IUserRepo
	nilID uuid.UUID
}

func (n nilWinnerUserFindRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if id == n.nilID {
		return nil, nil
	}
	return n.IUserRepo.FindByID(ctx, id)
}

func TestFinishGameServiceLoserNotFound(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p1, Username: "alice", Email: "a@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Game.Create(context.Background(), game); err != nil {
		t.Fatalf("create game: %v", err)
	}
	g := &GameUseCase{Games: repos.Game, Users: repos.User, MatchHistory: repos.MatchHistory, Repos: repos}
	err := finishGameService(context.Background(), g, game, p1, now)
	if err != nil {
		t.Fatalf("missing loser returns nil from mapRepoNotFound: %v", err)
	}
}

func TestFinishGameServiceInvalidWinnerOpponent(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	g := &GameUseCase{Games: repos.Game, Users: repos.User, MatchHistory: repos.MatchHistory, Repos: repos}
	err := finishGameService(context.Background(), g, game, uuid.New(), now)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("invalid winner opponent: %v", err)
	}
}

func TestFinishGameServiceFinishGameAlreadyFinished(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p1, Username: "alice", Email: "a@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create winner: %v", err)
	}
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p2, Username: "bob", Email: "b@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create loser: %v", err)
	}
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusFinished,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	g := &GameUseCase{Games: repos.Game, Users: repos.User, MatchHistory: repos.MatchHistory, Repos: repos}
	err := finishGameService(context.Background(), g, game, p1, now)
	if !errors.Is(err, ErrGameAlreadyFinished) {
		t.Fatalf("finish game on finished status: %v", err)
	}
}

func TestPurgeLogBatchesZeroSleepMultipleBatches(t *testing.T) {
	repos := openTestRepos(t)
	before := time.Now().UTC().AddDate(0, 0, -40)
	logTime := before.Add(-time.Hour)
	for i := 0; i < 2; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "old", Detail: []byte(`{}`), CreatedAt: logTime, UpdatedAt: logTime,
		}); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}
	if err := purgeLogBatches(context.Background(), before, 1, 0, repos.ActivityLog.DeleteOlderThan); err != nil {
		t.Fatalf("purge: %v", err)
	}
}

func TestIncrementUserWinCountFindError(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	p1 := uuid.New()
	if err := repos.User.Create(context.Background(), &model.User{
		ID: p1, Username: "alice", Email: "a@test.local", PasswordHash: "h",
		Role: model.RoleUser, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	base := repos.User
	repos.User = errWinCountUserFindRepo{IUserRepo: base, errID: p1}
	err := incrementUserWinCount(context.Background(), repos, p1, now)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("find error: %v", err)
	}
}

type errWinCountUserFindRepo struct {
	repository.IUserRepo
	errID uuid.UUID
}

func (e errWinCountUserFindRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if id == e.errID {
		return nil, context.Canceled
	}
	return e.IUserRepo.FindByID(ctx, id)
}

func TestPurgeLogBatchesContextCancelledDuringSleep(t *testing.T) {
	repos := openTestRepos(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	for i := 0; i < 3; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "ctx_cancel", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
		}); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	deleteFn := func(c context.Context, before time.Time, batchSize int) (int64, error) {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return 1, nil
	}
	err := purgeLogBatches(ctx, old, 1, 50*time.Millisecond, deleteFn)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled: %v", err)
	}
}

func TestPurgeLogBatchesDeleteFnError(t *testing.T) {
	old := time.Now().UTC().AddDate(0, 0, -40)
	err := purgeLogBatches(context.Background(), old, 100, time.Millisecond, func(context.Context, time.Time, int) (int64, error) {
		return 0, context.Canceled
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("delete fn error: %v", err)
	}
}

func TestPurgeLogBatchesSleepCompletes(t *testing.T) {
	repos := openTestRepos(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	for i := 0; i < 2; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "sleep", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
		}); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}
	if err := purgeLogBatches(context.Background(), old, 1, 10*time.Millisecond, repos.ActivityLog.DeleteOlderThan); err != nil {
		t.Fatalf("purge sleep: %v", err)
	}
}
