package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/numduel/numduel/db"
)

type SetupConfig struct {
	DatabaseURL       string
	BackupDatabaseURL string
	Migrate           bool
}

type SetupResult struct {
	Primary *gorm.DB
	Backup  *gorm.DB
	Repos   Repos
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
	result := &SetupResult{
		Primary: primary,
		Repos:   NewRepos(primary),
	}
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

func openDB(dsn string, migrate bool) (*gorm.DB, error) {
	gdb, err := db.Open(dsn)
	if err != nil {
		return nil, err
	}
	if migrate {
		if err := db.Migrate(gdb); err != nil {
			return nil, err
		}
	}
	return gdb, nil
}
