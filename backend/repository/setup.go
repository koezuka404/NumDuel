// 起動時の DB 初期化（接続・マイグレーション・バックアップ設定）
package repository

import (
	"context"
	"fmt"

	"github.com/numduel/numduel/db"
)

type SetupConfig struct {
	DatabaseURL       string
	BackupDatabaseURL string
	Migrate           bool // false のときマイグレーションをスキップ（server 起動時は migrate 済み想定）
}

type SetupResult struct {
	Primary *DB
	Backup  *DB
	Repo    IRepository
	Tx      TxManager
	Syncer  *BackupSyncer
}

func Setup(_ context.Context, cfg SetupConfig) (*SetupResult, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	primary, err := openDB(cfg.DatabaseURL, cfg.Migrate)
	if err != nil {
		return nil, fmt.Errorf("primary database: %w", err)
	}
	repo := NewRepository(primary)
	result := &SetupResult{Primary: primary, Repo: repo, Tx: NewTxManager(primary.Gorm())}
	if cfg.BackupDatabaseURL == "" {
		return result, nil
	}
	backup, err := openDB(cfg.BackupDatabaseURL, cfg.Migrate)
	if err != nil {
		return nil, fmt.Errorf("backup database: %w", err)
	}
	result.Backup = backup
	result.Syncer = NewBackupSyncer(primary, backup)
	return result, nil
}

func openDB(dsn string, migrate bool) (*DB, error) {
	gdb, err := db.Open(dsn)
	if err != nil {
		return nil, err
	}
	if migrate {
		if err := db.Migrate(gdb); err != nil {
			return nil, err
		}
	}
	return NewDB(gdb), nil
}
