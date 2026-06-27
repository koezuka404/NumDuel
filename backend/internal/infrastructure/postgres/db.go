package postgres

import (
	"context"
	"fmt"

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
	return &DB{gorm: gdb}, nil
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
	sql := `CREATE INDEX IF NOT EXISTS idx_games_active_players
ON games (player1_id, player2_id)
WHERE status IN ('WAITING_SECRET', 'IN_PROGRESS')`
	if err := d.gorm.Exec(sql).Error; err != nil {
		return fmt.Errorf("create partial index idx_games_active_players: %w", err)
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
