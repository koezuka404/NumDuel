package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config はアプリケーション設定。
type Config struct {
	DatabaseURL       string
	BackupDatabaseURL string
	Port              int

	MasterEmail    string
	MasterPassword string
}

// Load は環境変数から設定を読み込み、起動時バリデーションを行う。
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		BackupDatabaseURL: os.Getenv("BACKUP_DATABASE_URL"),
		MasterEmail:       os.Getenv("NUMDUEL_MASTER_EMAIL"),
		MasterPassword:    os.Getenv("NUMDUEL_MASTER_PASSWORD"),
		Port:              envInt("PORT", 8080),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("PORT must be positive")
	}
	return cfg, nil
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
