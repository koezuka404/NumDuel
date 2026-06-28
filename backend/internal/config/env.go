// 環境変数の読み込みと起動時バリデーション。
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	BackupDatabaseURL      string
	Port                   int
	MasterEmail            string
	MasterPassword         string
	JWTSecret              string
	JWTExpiryMinutes       int
	RefreshTokenExpiryDays int
	CookieSecure           bool // 本番は true（HTTPS 必須）
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		BackupDatabaseURL:      os.Getenv("BACKUP_DATABASE_URL"),
		MasterEmail:            os.Getenv("NUMDUEL_MASTER_EMAIL"),
		MasterPassword:         os.Getenv("NUMDUEL_MASTER_PASSWORD"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTExpiryMinutes:       envInt("JWT_EXPIRY_MINUTES", 60),
		RefreshTokenExpiryDays: envInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		CookieSecure:           envBool("COOKIE_SECURE", false),
		Port:                   envInt("PORT", 8080),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.Port <= 0 || cfg.JWTExpiryMinutes <= 0 || cfg.RefreshTokenExpiryDays <= 0 {
		return nil, fmt.Errorf("invalid numeric config")
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

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
