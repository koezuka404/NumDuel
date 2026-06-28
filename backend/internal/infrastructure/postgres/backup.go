// primary → backup への差分同期。Worker から定期実行する想定。
package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BackupSyncer struct {
	primary *DB
	backup  *DB
}

func NewBackupSyncer(primary, backup *DB) *BackupSyncer {
	return &BackupSyncer{primary: primary, backup: backup}
}

type SyncResult struct {
	SyncedRows int
	MaxUpdated time.Time
}

func (s *BackupSyncer) Sync(ctx context.Context, lastSyncedAt *time.Time) (SyncResult, error) {
	var result SyncResult
	tasks := []struct {
		name string
		run  func(context.Context, *time.Time) (int, time.Time, error)
	}{
		{"users", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, userModel{}, "id", func(m userModel) time.Time { return m.UpdatedAt })
		}},
		{"games", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, gameModel{}, "id", func(m gameModel) time.Time { return m.UpdatedAt })
		}},
		{"guesses", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, guessModel{}, "id", func(m guessModel) time.Time { return m.UpdatedAt })
		}},
		{"match_histories", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, matchHistoryModel{}, "id", func(m matchHistoryModel) time.Time { return m.UpdatedAt })
		}},
		{"rankings", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, rankingModel{}, "user_id", func(m rankingModel) time.Time { return m.UpdatedAt })
		}},
		{"activity_logs", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, activityLogModel{}, "id", func(m activityLogModel) time.Time { return m.UpdatedAt })
		}},
		{"login_logs", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary.gorm, s.backup.gorm, since, loginLogModel{}, "id", func(m loginLogModel) time.Time { return m.UpdatedAt })
		}},
	}
	for _, task := range tasks {
		n, maxUpdated, err := task.run(ctx, lastSyncedAt)
		if err != nil {
			return result, fmt.Errorf("sync %s: %w", task.name, err)
		}
		result.SyncedRows += n
		if maxUpdated.After(result.MaxUpdated) {
			result.MaxUpdated = maxUpdated
		}
	}
	return result, nil
}

func syncRows[T any](ctx context.Context, primary, backup *gorm.DB, since *time.Time, sample T, conflictCol string, updatedAt func(T) time.Time) (int, time.Time, error) {
	var rows []T
	q := primary.WithContext(ctx).Model(&sample)
	if since != nil {
		q = q.Where("updated_at > ?", *since)
	}
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		if t := updatedAt(row); t.After(maxUpdated) {
			maxUpdated = t
		}
	}
	err := backup.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: conflictCol}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}
