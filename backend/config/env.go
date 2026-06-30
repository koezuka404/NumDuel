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
	TurnDurationSeconds       int
	TurnTimeoutPollSeconds    int
	WSAllowedOrigins       []string
}

// Load は os.Getenv から設定を読み込む。
func Load() (*Config, error) {
	return LoadFromEnv(os.Getenv)
}

// LoadFromEnv は getenv から設定を読み込む（テスト用に注入可能）。
func LoadFromEnv(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		DatabaseURL:            getenv("DATABASE_URL"),
		BackupDatabaseURL:      getenv("BACKUP_DATABASE_URL"),
		JWTSecret:              getenv("JWT_SECRET"),
		JWTExpiryMinutes:       envInt(getenv, "JWT_EXPIRY_MINUTES", 60),
		RefreshTokenExpiryDays: envInt(getenv, "REFRESH_TOKEN_EXPIRY_DAYS", 7),
		CookieSecure:           envBool(getenv, "COOKIE_SECURE", false),
		Port:                   envInt(getenv, "PORT", 8080),
		GameSecretPepper:       getenv("GAME_SECRET_PEPPER"),
		GameLockSeconds:        envInt(getenv, "GAME_LOCK_SECONDS", 2),
		TurnDurationSeconds:    envInt(getenv, "TURN_DURATION_SECONDS", 30),
		TurnTimeoutPollSeconds: envInt(getenv, "TURN_TIMEOUT_POLL_SECONDS", 1),
	}
	if raw := getenv("WS_ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.WSAllowedOrigins = append(cfg.WSAllowedOrigins, o)
			}
		}
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (set in .env or export in shell)")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required (set in .env or export in shell)")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.GameSecretPepper == "" {
		return nil, fmt.Errorf("GAME_SECRET_PEPPER is required (set in .env or export in shell)")
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

func (c *Config) TurnTimeoutPollInterval() time.Duration {
	return time.Duration(c.TurnTimeoutPollSeconds) * time.Second
}

func envInt(getenv func(string) string, key string, fallback int) int {
	raw := getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func envBool(getenv func(string) string, key string, fallback bool) bool {
	raw := getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
