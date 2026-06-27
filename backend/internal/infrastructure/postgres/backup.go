package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BackupSyncer は本番 DB からバックアップ DB へ差分 UPSERT する。
type BackupSyncer struct {
	primary *DB
	backup  *DB
}

// NewBackupSyncer は BackupSyncer を生成する。
func NewBackupSyncer(primary, backup *DB) *BackupSyncer {
	return &BackupSyncer{primary: primary, backup: backup}
}

// SyncResult は 1 回の同期結果。
type SyncResult struct {
	SyncedRows int
	MaxUpdated time.Time
}

// Sync は updated_at > lastSyncedAt の行をバックアップ DB へ UPSERT する。
// lastSyncedAt が nil のときは全件同期する。
func (s *BackupSyncer) Sync(ctx context.Context, lastSyncedAt *time.Time) (SyncResult, error) {
	var result SyncResult

	tables := []struct {
		name string
		fn   func(context.Context, *time.Time) (int, time.Time, error)
	}{
		{"users", s.syncUsers},
		{"games", s.syncGames},
		{"guesses", s.syncGuesses},
		{"match_histories", s.syncMatchHistories},
		{"rankings", s.syncRankings},
		{"activity_logs", s.syncActivityLogs},
		{"login_logs", s.syncLoginLogs},
	}

	for _, table := range tables {
		n, maxUpdated, err := table.fn(ctx, lastSyncedAt)
		if err != nil {
			return result, fmt.Errorf("sync %s: %w", table.name, err)
		}
		result.SyncedRows += n
		if maxUpdated.After(result.MaxUpdated) {
			result.MaxUpdated = maxUpdated
		}
	}
	return result, nil
}

func applySince(q *gorm.DB, lastSyncedAt *time.Time) *gorm.DB {
	if lastSyncedAt == nil {
		return q
	}
	return q.Where("updated_at > ?", *lastSyncedAt)
}

func maxTime(current, candidate time.Time) time.Time {
	if candidate.After(current) {
		return candidate
	}
	return current
}

func (s *BackupSyncer) syncUsers(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []userModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&userModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncGames(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []gameModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&gameModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncGuesses(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []guessModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&guessModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncMatchHistories(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []matchHistoryModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&matchHistoryModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncRankings(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []rankingModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&rankingModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncActivityLogs(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []activityLogModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&activityLogModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}

func (s *BackupSyncer) syncLoginLogs(ctx context.Context, lastSyncedAt *time.Time) (int, time.Time, error) {
	var rows []loginLogModel
	q := applySince(s.primary.gorm.WithContext(ctx).Model(&loginLogModel{}), lastSyncedAt)
	if err := q.Find(&rows).Error; err != nil {
		return 0, time.Time{}, err
	}
	if len(rows) == 0 {
		return 0, time.Time{}, nil
	}
	maxUpdated := time.Time{}
	for _, row := range rows {
		maxUpdated = maxTime(maxUpdated, row.UpdatedAt)
	}
	err := s.backup.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&rows).Error
	return len(rows), maxUpdated, err
}
