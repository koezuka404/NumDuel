// 環境変数の読み込みと起動時バリデーション
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
	AdminLockSeconds       int // 管理操作 Redis ロック TTL（§13.10.2、既定 5）
	TurnDurationSeconds        int
	TurnTimeoutPollSeconds     int
	SecretSetupSeconds         int
	SecretTimeoutPollSeconds   int
	SessionTimeoutMinutes      int // 無操作とみなす分数（デフォルト 5）
	AutoLogoutPollSeconds      int // AutoLogoutWorker のポーリング間隔秒（デフォルト 60）
	BackupCron                 string // BackupWorker スケジュール（§12.8、既定 03:00 UTC）
	RankingRebuildCron         string // RankingRebuildWorker スケジュール（§12.6、既定 10分毎 UTC）
	LogRetentionCron           string // LogRetentionWorker スケジュール（§12.7、既定 日曜 03:30 UTC）
	ActivityLogRetentionDays   int
	LoginLogRetentionDays      int
	WSLogRetentionDays         int
	RetentionBatchSize         int
	RetentionBatchSleepMs      int
	RefreshTokenCleanupCron    string
	RefreshTokenCleanupGraceDays int
	CORSAllowedOrigins         []string
	WSAllowedOrigins           []string
	MasterEmail                string
	MasterPassword             string
}

// Load は os.Getenv から設定を読み込む
func Load() (*Config, error) {
	return LoadFromEnv(os.Getenv)
}

// LoadFromEnv は getenv から設定を読み込む（テスト用に注入可能）
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
		AdminLockSeconds:       envInt(getenv, "ADMIN_LOCK_SECONDS", 5),
		TurnDurationSeconds:    envInt(getenv, "TURN_DURATION_SECONDS", 30),
		TurnTimeoutPollSeconds:   envInt(getenv, "TURN_TIMEOUT_POLL_SECONDS", 1),
		SecretSetupSeconds:       envInt(getenv, "SECRET_SETUP_SECONDS", 60),
		SecretTimeoutPollSeconds: envInt(getenv, "SECRET_TIMEOUT_POLL_SECONDS", 1),
		SessionTimeoutMinutes:    envInt(getenv, "SESSION_TIMEOUT_MINUTES", 5),
		AutoLogoutPollSeconds:    envInt(getenv, "AUTO_LOGOUT_POLL_SECONDS", 60),
		BackupCron:               envString(getenv, "BACKUP_CRON", "0 3 * * *"),
		RankingRebuildCron:       envString(getenv, "RANKING_REBUILD_CRON", "*/10 * * * *"),
		LogRetentionCron:         envString(getenv, "LOG_RETENTION_CRON", "30 3 * * 0"),
		ActivityLogRetentionDays: envInt(getenv, "ACTIVITY_LOG_RETENTION_DAYS", 90),
		LoginLogRetentionDays:    envInt(getenv, "LOGIN_LOG_RETENTION_DAYS", 90),
		WSLogRetentionDays:       envInt(getenv, "WS_LOG_RETENTION_DAYS", 30),
		RetentionBatchSize:           envInt(getenv, "RETENTION_BATCH_SIZE", 1000),
		RetentionBatchSleepMs:        envInt(getenv, "RETENTION_BATCH_SLEEP_MS", 100),
		RefreshTokenCleanupCron:      envString(getenv, "REFRESH_TOKEN_CLEANUP_CRON", "0 4 * * *"),
		RefreshTokenCleanupGraceDays: envInt(getenv, "REFRESH_TOKEN_CLEANUP_GRACE_DAYS", 7),
		MasterEmail:              getenv("NUMDUEL_MASTER_EMAIL"),
		MasterPassword:           getenv("NUMDUEL_MASTER_PASSWORD"),
	}
	if raw := getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, o)
			}
		}
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
	if cfg.GameLockSeconds <= 0 || cfg.AdminLockSeconds <= 0 || cfg.TurnDurationSeconds <= 0 || cfg.SecretSetupSeconds <= 0 {
		return nil, fmt.Errorf("invalid game timing config")
	}
	if cfg.SessionTimeoutMinutes <= 0 || cfg.AutoLogoutPollSeconds <= 0 {
		return nil, fmt.Errorf("invalid session config")
	}
	return cfg, nil
}

func (c *Config) GameLockTTL() time.Duration {
	return time.Duration(c.GameLockSeconds) * time.Second
}

func (c *Config) AdminLockTTL() time.Duration {
	return time.Duration(c.AdminLockSeconds) * time.Second
}

func (c *Config) TurnDuration() time.Duration {
	return time.Duration(c.TurnDurationSeconds) * time.Second
}

func (c *Config) TurnTimeoutPollInterval() time.Duration {
	return time.Duration(c.TurnTimeoutPollSeconds) * time.Second
}

func (c *Config) SecretSetupDuration() time.Duration {
	return time.Duration(c.SecretSetupSeconds) * time.Second
}

func (c *Config) SecretTimeoutPollInterval() time.Duration {
	return time.Duration(c.SecretTimeoutPollSeconds) * time.Second
}

func (c *Config) SessionTimeout() time.Duration {
	return time.Duration(c.SessionTimeoutMinutes) * time.Minute
}

func (c *Config) AutoLogoutPollInterval() time.Duration {
	return time.Duration(c.AutoLogoutPollSeconds) * time.Second
}

func (c *Config) RetentionBatchSleep() time.Duration {
	return time.Duration(c.RetentionBatchSleepMs) * time.Millisecond
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

func envString(getenv func(string) string, key, fallback string) string {
	raw := getenv(key)
	if raw == "" {
		return fallback
	}
	return raw
}
