package postgres

import (
	"context"
	"fmt"

	"github.com/numduel/numduel/internal/domain"
)

// SetupConfig は DB 初期化に必要な設定。
type SetupConfig struct {
	DatabaseURL       string
	BackupDatabaseURL string
	MasterEmail       string
	MasterPassword    string
}

// SetupResult は初期化後の接続と Repository。
type SetupResult struct {
	Primary *DB
	Backup  *DB
	Repo    domain.Repository
	Syncer  *BackupSyncer
}

// Setup は接続・マイグレーション・インデックス・master seed を行う。
func Setup(ctx context.Context, cfg SetupConfig) (*SetupResult, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	primary, err := Open(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open primary database: %w", err)
	}
	if err := primary.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping primary database: %w", err)
	}
	if err := primary.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migrate primary database: %w", err)
	}
	if err := primary.CreateIndexes(); err != nil {
		return nil, fmt.Errorf("create primary indexes: %w", err)
	}

	repo := NewRepository(primary)
	if err := SeedMaster(ctx, repo, cfg.MasterEmail, cfg.MasterPassword); err != nil {
		return nil, fmt.Errorf("seed master user: %w", err)
	}

	result := &SetupResult{
		Primary: primary,
		Repo:    repo,
	}

	if cfg.BackupDatabaseURL == "" {
		return result, nil
	}

	backup, err := Open(cfg.BackupDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open backup database: %w", err)
	}
	if err := backup.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping backup database: %w", err)
	}
	if err := backup.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migrate backup database: %w", err)
	}
	if err := backup.CreateIndexes(); err != nil {
		return nil, fmt.Errorf("create backup indexes: %w", err)
	}

	result.Backup = backup
	result.Syncer = NewBackupSyncer(primary, backup)
	return result, nil
}
