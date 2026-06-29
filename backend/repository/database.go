// PostgreSQL 接続・マイグレーション・トランザクション管理。
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/numduel/numduel/model"
)

type DB struct {
	gorm *gorm.DB
}

func Open(dsn string) (*DB, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	configurePool(gdb)
	return &DB{gorm: gdb}, nil
}

func configurePool(gdb *gorm.DB) {
	sqlDB, err := gdb.DB()
	if err != nil {
		return
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func (d *DB) Ping(ctx context.Context) error {
	sqlDB, err := d.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *DB) Gorm() *gorm.DB { return d.gorm }

func (d *DB) AutoMigrate() error {
	return d.gorm.AutoMigrate(model.MigrateTargets()...)
}

func (d *DB) CreateIndexes() error {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_games_active_players
ON games (player1_id, player2_id)
WHERE status IN ('WAITING_SECRET', 'IN_PROGRESS')`,
		`CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_games_updated_at ON games (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_guesses_updated_at ON guesses (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_match_histories_updated_at ON match_histories (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_rankings_updated_at ON rankings (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_updated_at ON activity_logs (updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_updated_at ON login_logs (updated_at)`,
	}
	for _, sql := range indexes {
		if err := d.gorm.Exec(sql).Error; err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}
