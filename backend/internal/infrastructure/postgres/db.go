package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/numduel/numduel/internal/domain"
)

type DB struct {
	gorm *gorm.DB
}

type gormTx struct {
	tx *gorm.DB
}

func (t gormTx) db() *gorm.DB { return t.tx }

// Open は PostgreSQL へ接続する。
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

// Ping は DB 接続を確認する。
func (d *DB) Ping(ctx context.Context) error {
	sqlDB, err := d.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *DB) Gorm() *gorm.DB { return d.gorm }

// AutoMigrate は全テーブルを作成・更新する。
func (d *DB) AutoMigrate() error {
	return d.gorm.AutoMigrate(
		&userModel{},
		&gameModel{},
		&guessModel{},
		&matchHistoryModel{},
		&rankingModel{},
		&matchingQueueModel{},
		&activityLogModel{},
		&loginLogModel{},
		&wsConnectionLogModel{},
		&refreshTokenModel{},
	)
}

// CreateIndexes は AutoMigrate では作成できないインデックスを追加する。
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

func (d *DB) Begin(ctx context.Context) (domain.Transaction, error) {
	tx := d.gorm.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return gormTx{tx: tx}, nil
}

func (d *DB) Commit(tx domain.Transaction) error {
	gtx, ok := tx.(gormTx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return gtx.tx.Commit().Error
}

func (d *DB) Rollback(tx domain.Transaction) error {
	gtx, ok := tx.(gormTx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return gtx.tx.Rollback().Error
}

func forUpdate(db *gorm.DB) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

func dbOrGlobal(db *gorm.DB, tx domain.Transaction) (*gorm.DB, error) {
	if tx == nil {
		return db, nil
	}
	gtx, ok := tx.(gormTx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type")
	}
	return gtx.db(), nil
}

// OpenSQLite はテスト用 SQLite 接続を開く。
func OpenSQLite(dsn string) (*DB, error) {
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return &DB{gorm: gdb}, nil
}
