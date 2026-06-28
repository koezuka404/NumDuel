// 起動時の DB 初期化（接続・マイグレーション・シード・バックアップ設定）。
package postgres

import (
	"context"
	"fmt"

	"github.com/numduel/numduel/internal/domain"
)

type SetupConfig struct {
	DatabaseURL       string
	BackupDatabaseURL string
	MasterEmail       string
	MasterPassword    string
}

type SetupResult struct {
	Primary *DB
	Backup  *DB
	Repo    domain.Repository
	Syncer  *BackupSyncer
}

// Setup は primary DB を準備し、master シードと Repository を返す。
func Setup(ctx context.Context, cfg SetupConfig) (*SetupResult, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	primary, err := openReadyDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("primary database: %w", err)
	}
	repo := NewRepository(primary)
	if err := SeedMaster(ctx, repo, cfg.MasterEmail, cfg.MasterPassword); err != nil {
		return nil, fmt.Errorf("seed master: %w", err)
	}
	result := &SetupResult{Primary: primary, Repo: repo}
	if cfg.BackupDatabaseURL == "" {
		return result, nil
	}
	backup, err := openReadyDB(ctx, cfg.BackupDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("backup database: %w", err)
	}
	result.Backup = backup
	result.Syncer = NewBackupSyncer(primary, backup)
	return result, nil
}

// openReadyDB は接続 → Ping → AutoMigrate → インデックス作成まで行う。
func openReadyDB(ctx context.Context, dsn string) (*DB, error) {
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	if err := db.CreateIndexes(); err != nil {
		return nil, err
	}
	return db, nil
}
