package usecase

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/repository"
)

const (
	defaultRetentionBatchSize = 1000
	defaultRetentionBatchSleep = 100 * time.Millisecond
)

// LogRetentionDeps は LogRetentionWorker / RunLogRetention の依存
type LogRetentionDeps struct {
	Repo                     repository.Repos
	ActivityLogRetentionDays int
	LoginLogRetentionDays    int
	WSLogRetentionDays       int
	BatchSize                int
	BatchSleep               time.Duration
	Now                      func() time.Time
}

// RunLogRetention は保持期限を超えたログをバッチ削除する（§12.7）
// テーブル単位の失敗は警告ログに記録し、次回 Worker に持ち越す
func RunLogRetention(ctx context.Context, d LogRetentionDeps) {
	if d.Repo == nil {
		return
	}
	now := d.now()
	batchSize := d.BatchSize
	if batchSize <= 0 {
		batchSize = defaultRetentionBatchSize
	}
	sleep := d.BatchSleep
	if sleep <= 0 {
		sleep = defaultRetentionBatchSleep
	}

	type task struct {
		name   string
		before time.Time
		delete func(context.Context, time.Time, int) (int64, error)
	}
	var tasks []task
	if d.ActivityLogRetentionDays > 0 {
		before := now.AddDate(0, 0, -d.ActivityLogRetentionDays)
		tasks = append(tasks, task{
			name: "activity_logs", before: before,
			delete: d.Repo.ActivityLog.DeleteOlderThan,
		})
	}
	if d.LoginLogRetentionDays > 0 {
		before := now.AddDate(0, 0, -d.LoginLogRetentionDays)
		tasks = append(tasks, task{
			name: "login_logs", before: before,
			delete: d.Repo.LoginLog.DeleteOlderThan,
		})
	}
	if d.WSLogRetentionDays > 0 {
		before := now.AddDate(0, 0, -d.WSLogRetentionDays)
		tasks = append(tasks, task{
			name: "ws_connection_logs", before: before,
			delete: d.Repo.WSConnectionLog.DeleteOlderThan,
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

func (d LogRetentionDeps) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}
