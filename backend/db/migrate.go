package db

import (
	"fmt"

	"github.com/numduel/numduel/model"

	"gorm.io/gorm"
)

func execSQL(gdb *gorm.DB, q string) error {
	return gdb.Exec(q).Error
}

var execSQLFn = execSQL

// Migrate はスキーマの AutoMigrate と追加インデックスを適用する
func Migrate(gdb *gorm.DB) error {
	if err := gdb.AutoMigrate(model.MigrateTargets()...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
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
		if err := execSQLFn(gdb, sql); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}
