// 起動時の DB 初期化（接続・マイグレーション・バックアップ設定）。
package repository

import (
	"context"
	"fmt"
)

type SetupConfig struct {
	DatabaseURL       string
	BackupDatabaseURL string
}

type SetupResult struct {
	Primary *DB
	Backup  *DB
	Repo    IRepository
	Tx      TxManager
	Syncer  *BackupSyncer
}

func Setup(ctx context.Context, cfg SetupConfig) (*SetupResult, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	primary, err := openReadyDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("primary database: %w", err)
	}
	repo := NewRepository(primary)
	result := &SetupResult{Primary: primary, Repo: repo, Tx: NewTxManager(primary.Gorm())}
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

func openReadyDB(ctx context.Context, dsn string) (*DB, error) {
	conn, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}
	if err := conn.AutoMigrate(); err != nil {
		return nil, err
	}
	if err := conn.CreateIndexes(); err != nil {
		return nil, err
	}
	return conn, nil
}
