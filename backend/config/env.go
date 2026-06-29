// 環境変数の読み込みと起動時バリデーション。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL            string
	BackupDatabaseURL      string
	Port                   int
	JWTSecret              string
	JWTExpiryMinutes       int
	RefreshTokenExpiryDays int
	CookieSecure           bool
	GameSecretPepper       string
	GameLockSeconds        int
	TurnDurationSeconds    int
	WSAllowedOrigins       []string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		BackupDatabaseURL:      os.Getenv("BACKUP_DATABASE_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTExpiryMinutes:       envInt("JWT_EXPIRY_MINUTES", 60),
		RefreshTokenExpiryDays: envInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		CookieSecure:           envBool("COOKIE_SECURE", false),
		Port:                   envInt("PORT", 8080),
		GameSecretPepper:       os.Getenv("GAME_SECRET_PEPPER"),
		GameLockSeconds:        envInt("GAME_LOCK_SECONDS", 2),
		TurnDurationSeconds:    envInt("TURN_DURATION_SECONDS", 30),
	}
	if raw := os.Getenv("WS_ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.WSAllowedOrigins = append(cfg.WSAllowedOrigins, o)
			}
		}
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
	if cfg.GameSecretPepper == "" {
		return nil, fmt.Errorf("GAME_SECRET_PEPPER is required")
	}
	if len([]byte(cfg.GameSecretPepper)) < 32 {
		return nil, fmt.Errorf("GAME_SECRET_PEPPER must be at least 32 bytes")
	}
	if cfg.Port <= 0 || cfg.JWTExpiryMinutes <= 0 || cfg.RefreshTokenExpiryDays <= 0 {
		return nil, fmt.Errorf("invalid numeric config")
	}
	if cfg.GameLockSeconds <= 0 || cfg.TurnDurationSeconds <= 0 {
		return nil, fmt.Errorf("invalid game timing config")
	}
	return cfg, nil
}

func (c *Config) GameLockTTL() time.Duration {
	return time.Duration(c.GameLockSeconds) * time.Second
}

func (c *Config) TurnDuration() time.Duration {
	return time.Duration(c.TurnDurationSeconds) * time.Second
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
