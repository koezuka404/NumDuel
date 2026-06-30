package usecase

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/repository"
)

//ログ保持期間のクリーンアップユースケース。
type ILogRetentionUsecase interface {
	Run(ctx context.Context)
}

const (
	defaultRetentionBatchSize  = 1000
	defaultRetentionBatchSleep = 100 * time.Millisecond
)

type LogRetentionUseCase struct {
	ActivityLogs repository.IActivityLogRepo
	LoginLogs    repository.ILoginLogRepo
	WSLogs       repository.IWSConnectionLogRepo
	ActivityDays int
	LoginDays    int
	WSDays       int
	BatchSize    int
	BatchSleep   time.Duration
	Now          func() time.Time
}

func (l *LogRetentionUseCase) now() time.Time {
	if l != nil && l.Now != nil {
		return l.Now().UTC()
	}
	return time.Now().UTC()
}

func (l *LogRetentionUseCase) Run(ctx context.Context) {
	now := l.now()
	batchSize := l.BatchSize
	if batchSize <= 0 {
		batchSize = defaultRetentionBatchSize
	}
	sleep := l.BatchSleep
	if sleep <= 0 {
		sleep = defaultRetentionBatchSleep
	}
	type task struct {
		name   string
		before time.Time
		delete func(context.Context, time.Time, int) (int64, error)
	}
	var tasks []task
	if l.ActivityDays > 0 {
		tasks = append(tasks, task{
			name: "activity_logs", before: now.AddDate(0, 0, -l.ActivityDays),
			delete: l.ActivityLogs.DeleteOlderThan,
		})
	}
	if l.LoginDays > 0 {
		tasks = append(tasks, task{
			name: "login_logs", before: now.AddDate(0, 0, -l.LoginDays),
			delete: l.LoginLogs.DeleteOlderThan,
		})
	}
	if l.WSDays > 0 {
		tasks = append(tasks, task{
			name: "ws_connection_logs", before: now.AddDate(0, 0, -l.WSDays),
			delete: l.WSLogs.DeleteOlderThan,
		})
	}
	for _, t := range tasks {
		if err := purgeLogBatches(ctx, t.before, batchSize, sleep, t.delete); err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			log.Printf("log retention: %s: %v", t.name, err)
		}
	}
}

func NewLogRetentionUseCase(repos repository.Repos, activityDays, loginDays, wsDays, batchSize int, batchSleep time.Duration) *LogRetentionUseCase {
	return &LogRetentionUseCase{
		ActivityLogs: repos.ActivityLog,
		LoginLogs:    repos.LoginLog,
		WSLogs:       repos.WSConnectionLog,
		ActivityDays: activityDays,
		LoginDays:    loginDays,
		WSDays:       wsDays,
		BatchSize:    batchSize,
		BatchSleep:   batchSleep,
	}
}

func purgeLogBatches(
	ctx context.Context,
	before time.Time,
	batchSize int,
	sleep time.Duration,
	deleteFn func(context.Context, time.Time, int) (int64, error),
) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := deleteFn(ctx, before, batchSize)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		if sleep <= 0 {
			continue
		}
		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
