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
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

type errListGamesBeforeRepo struct {
	repository.IGameRepo
}

func (errListGamesBeforeRepo) ListByStatusCreatedBefore(context.Context, model.GameStatus, time.Time) ([]*model.Game, error) {
	return nil, context.Canceled
}

type nilEntryGamesBeforeRepo struct {
	repository.IGameRepo
}

func (nilEntryGamesBeforeRepo) ListByStatusCreatedBefore(context.Context, model.GameStatus, time.Time) ([]*model.Game, error) {
	return []*model.Game{nil}, nil
}

type errGameFindForUpdateRepo struct {
	repository.IGameRepo
	failID uuid.UUID
}

func (e errGameFindForUpdateRepo) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	if id == e.failID {
		return nil, context.Canceled
	}
	return e.IGameRepo.FindByIDForUpdate(ctx, id)
}

type errLockStore struct{}

func (errLockStore) AcquireLock(context.Context, string, time.Duration) (bool, error) {
	return false, context.Canceled
}

func TestAutoLogoutWorkerTickError(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	uc := usecase.NewAutoLogoutUseCase(repos, nil, nil, time.Hour)
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	w := &AutoLogoutWorker{AutoLogout: uc}
	w.tick(context.Background(), time.Now().UTC())
}

func TestBackupWorkerCronSyncError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := infrredis.NewStore(rdb)
	mr.Close()
	_ = rdb.Close()

	primaryGDB, _ := testutil.OpenSQLiteDB(t)
	syncer := repository.NewBackupSyncer(primaryGDB, primaryGDB)
	backupUC := usecase.NewBackupUseCase(syncer, store, 1)
	runCronWorker(t, (&BackupWorker{Backup: backupUC, Cron: "@every 1s"}).Run)
}

func TestRankingRebuildWorkerCronRebuildError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	ranking.Locks = errLockStore{}
	runCronWorker(t, (&RankingRebuildWorker{Ranking: ranking, Cron: "@every 1s"}).Run)
}

func TestSecretSetupTimeoutWorkerTickZeroDuration(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.SecretSetup = 0

	w := &SecretSetupTimeoutWorker{Game: gameUC}
	w.tick(t.Context(), time.Now().UTC())
}

func TestSecretSetupTimeoutWorkerTickListError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.SecretSetup = time.Minute
	gameUC.Games = errListGamesBeforeRepo{IGameRepo: repos.Game}

	w := &SecretSetupTimeoutWorker{Game: gameUC}
	w.tick(t.Context(), time.Now().UTC())
}

func TestSecretSetupTimeoutWorkerTickSkipsNilGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.SecretSetup = time.Minute
	gameUC.Games = nilEntryGamesBeforeRepo{IGameRepo: repos.Game}

	w := &SecretSetupTimeoutWorker{Game: gameUC}
	w.tick(t.Context(), time.Now().UTC())
}

func TestSecretSetupTimeoutWorkerTickCancelError(t *testing.T) {
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

	gameUC.Games = errGameFindForUpdateRepo{IGameRepo: repos.Game, failID: gameID}

	w := &SecretSetupTimeoutWorker{Game: gameUC}
	w.tick(t.Context(), now)
}
