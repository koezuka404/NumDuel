package worker

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	infrredis "github.com/numduel/numduel/redis"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func newRedisStore(t *testing.T) *infrredis.Store {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return infrredis.NewStore(rdb)
}

func matchTwoPlayers(t *testing.T, match *usecase.MatchingUseCase, a, b uuid.UUID) uuid.UUID {
	t.Helper()
	if _, err := match.Start(t.Context(), a); err != nil {
		t.Fatalf("start a: %v", err)
	}
	if _, err := match.Start(t.Context(), b); err != nil {
		t.Fatalf("start b: %v", err)
	}
	status, err := match.Status(t.Context(), a)
	if err != nil || status.GameID == nil {
		t.Fatalf("status: %+v err=%v", status, err)
	}
	return *status.GameID
}

func setBothSecrets(t *testing.T, gameUC *usecase.GameUseCase, gameID, a, b uuid.UUID, secretA, secretB string) {
	t.Helper()
	if err := gameUC.SetSecretNumber(t.Context(), a, gameID, secretA); err != nil {
		t.Fatalf("secret a: %v", err)
	}
	if err := gameUC.SetSecretNumber(t.Context(), b, gameID, secretB); err != nil {
		t.Fatalf("secret b: %v", err)
	}
}

func TestTurnTimeoutWorkerTickHandlesExpiredTurn(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	store := newRedisStore(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = store

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoPlayers(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, err := repos.Game.FindByID(t.Context(), gameID)
	if err != nil {
		t.Fatalf("find game: %v", err)
	}
	now := time.Now().UTC()
	if err := store.SetTurn(t.Context(), gameID, game.CurrentTurn, a.ID, now.Add(-time.Minute), now.Add(-time.Second)); err != nil {
		t.Fatalf("set turn: %v", err)
	}

	w := &TurnTimeoutWorker{Store: store, Game: gameUC}
	w.tick(t.Context(), now)

	guesses, err := repos.Guess.ListByGameAndPlayer(t.Context(), gameID, a.ID)
	if err != nil || len(guesses) == 0 || !guesses[0].IsAuto {
		t.Fatalf("auto guess expected: %+v err=%v", guesses, err)
	}
}

func TestTurnTimeoutWorkerTickListError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := infrredis.NewStore(rdb)
	mr.Close()
	_ = rdb.Close()

	w := &TurnTimeoutWorker{Store: store, Game: testutil.NewGameUC(t, repos)}
	w.tick(context.Background(), time.Now().UTC())
}

func TestTurnTimeoutWorkerRun(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	store := newRedisStore(t)
	gameUC := testutil.NewGameUC(t, repos)
	w := &TurnTimeoutWorker{Store: store, Game: gameUC, Interval: time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)
}

func runCronWorker(t *testing.T, run func(context.Context)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1200 * time.Millisecond):
		cancel()
		<-done
	}
}

func TestBackupWorkerRunsCronJob(t *testing.T) {
	backupUC := usecase.NewBackupUseCase(nil, nil, 0)
	runCronWorker(t, (&BackupWorker{Backup: backupUC, Cron: "@every 1s"}).Run)
}

func TestLogRetentionWorkerRunsCronJob(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	retention := usecase.NewLogRetentionUseCase(repos, 0, 0, 0, 100, 0)
	runCronWorker(t, (&LogRetentionWorker{Retention: retention, Cron: "@every 1s"}).Run)
}

func TestRankingRebuildWorkerRunsCronJob(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	runCronWorker(t, (&RankingRebuildWorker{Ranking: ranking, Cron: "@every 1s"}).Run)
}

func TestRefreshTokenCleanupWorkerRunsCronJob(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	runCronWorker(t, (&RefreshTokenCleanupWorker{Auth: auth, Cron: "@every 1s"}).Run)
}

func TestSecretSetupTimeoutWorkerTickWithExpiredGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.SecretSetup = time.Minute

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoPlayers(t, match, a.ID, b.ID)

	now := time.Now().UTC()
	game, err := repos.Game.FindByID(t.Context(), gameID)
	if err != nil {
		t.Fatalf("find game: %v", err)
	}
	game.CreatedAt = now.Add(-2 * time.Minute)
	if err := repos.Game.Update(t.Context(), game); err != nil {
		t.Fatalf("update game: %v", err)
	}

	w := &SecretSetupTimeoutWorker{Game: gameUC, Interval: time.Millisecond}
	w.tick(t.Context(), now)

	updated, err := repos.Game.FindByID(t.Context(), gameID)
	if err != nil || updated.Status != model.GameStatusFinished {
		t.Fatalf("game cancelled: %+v err=%v", updated, err)
	}
}

func TestTurnTimeoutWorkerTickHandleTimeoutError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	store := newRedisStore(t)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = store

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	match := testutil.NewMatchingUC(repos)
	gameID := matchTwoPlayers(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	now := time.Now().UTC()
	wrongPlayer := uuid.New()
	if err := store.SetTurn(t.Context(), gameID, 1, wrongPlayer, now.Add(-time.Minute), now.Add(-time.Second)); err != nil {
		t.Fatalf("set turn: %v", err)
	}

	w := &TurnTimeoutWorker{Store: store, Game: gameUC}
	w.tick(t.Context(), now)
}

func TestCronWorkersInvalidCron(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, repos := testutil.OpenSQLiteDB(t)
	(&LogRetentionWorker{Retention: usecase.NewLogRetentionUseCase(repos, 0, 0, 0, 100, 0), Cron: "bad"}).Run(ctx)
	(&RankingRebuildWorker{Ranking: testutil.NewRankingUC(repos), Cron: "bad"}).Run(ctx)
	(&RefreshTokenCleanupWorker{Auth: testutil.NewAuthUC(t, repos), Cron: "bad"}).Run(ctx)
}
