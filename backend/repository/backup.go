package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/numduel/numduel/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BackupSyncer struct {
	primary *gorm.DB
	backup  *gorm.DB
}

func NewBackupSyncer(primary, backup *gorm.DB) *BackupSyncer {
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
			return syncRows(ctx, s.primary, s.backup, since, model.User{}, "id", func(m model.User) time.Time { return m.UpdatedAt })
		}},
		{"games", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.Game{}, "id", func(m model.Game) time.Time { return m.UpdatedAt })
		}},
		{"guesses", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.Guess{}, "id", func(m model.Guess) time.Time { return m.UpdatedAt })
		}},
		{"match_histories", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.MatchHistory{}, "id", func(m model.MatchHistory) time.Time { return m.UpdatedAt })
		}},
		{"rankings", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.Ranking{}, "user_id", func(m model.Ranking) time.Time { return m.UpdatedAt })
		}},
		{"activity_logs", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.ActivityLog{}, "id", func(m model.ActivityLog) time.Time { return m.UpdatedAt })
		}},
		{"login_logs", func(ctx context.Context, since *time.Time) (int, time.Time, error) {
			return syncRows(ctx, s.primary, s.backup, since, model.LoginLog{}, "id", func(m model.LoginLog) time.Time { return m.UpdatedAt })
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
