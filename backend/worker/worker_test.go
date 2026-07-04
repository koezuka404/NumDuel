package worker

import (
	"context"
	"testing"
	"time"

	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestWorkersRunEarlyReturn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	(&AutoLogoutWorker{}).Run(ctx)
	(&AutoLogoutWorker{Interval: time.Second}).Run(ctx)
	(&TurnTimeoutWorker{}).Run(ctx)
	(&TurnTimeoutWorker{Interval: time.Second}).Run(ctx)
	(&SecretSetupTimeoutWorker{}).Run(ctx)
	(&SecretSetupTimeoutWorker{Interval: time.Second}).Run(ctx)
	(&BackupWorker{}).Run(ctx)
	(&BackupWorker{Cron: "* * * * * *"}).Run(ctx)
	(&RankingRebuildWorker{}).Run(ctx)
	(&LogRetentionWorker{}).Run(ctx)
	(&RefreshTokenCleanupWorker{}).Run(ctx)
}

func TestAutoLogoutWorkerTick(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	uc := usecase.NewAutoLogoutUseCase(repos, nil, nil, time.Hour)
	w := &AutoLogoutWorker{AutoLogout: uc, Interval: time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)
}

func TestSecretSetupTimeoutWorkerTick(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	w := &SecretSetupTimeoutWorker{Game: gameUC, Interval: time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)
}

func TestBackupWorkerInvalidCron(t *testing.T) {
	backupUC := usecase.NewBackupUseCase(nil, nil, 0)
	w := &BackupWorker{Backup: backupUC, Cron: "not a cron"}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	w.Run(ctx)
}
