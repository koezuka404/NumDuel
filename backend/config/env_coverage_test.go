package config

import (
	"strings"
	"testing"
)

func TestLoadFromEnvMissingJWTSecret(t *testing.T) {
	_, err := LoadFromEnv(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgres://localhost/numduel"
		case "GAME_SECRET_PEPPER":
			return "abcdefghijklmnopqrstuvwxyz1234567890abcd"
		default:
			return ""
		}
	})
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET is required") {
		t.Fatalf("expected JWT_SECRET required, got %v", err)
	}
}

func TestLoadFromEnvMissingPepper(t *testing.T) {
	_, err := LoadFromEnv(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgres://localhost/numduel"
		case "JWT_SECRET":
			return "abcdefghijklmnopqrstuvwxyz123456"
		default:
			return ""
		}
	})
	if err == nil || !strings.Contains(err.Error(), "GAME_SECRET_PEPPER is required") {
		t.Fatalf("expected pepper required, got %v", err)
	}
}

func TestLoadFromEnvInvalidRefreshTokenDays(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "REFRESH_TOKEN_EXPIRY_DAYS" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid numeric") {
		t.Fatalf("expected numeric error, got %v", err)
	}
}

func TestLoadFromEnvInvalidAutoLogoutPoll(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "AUTO_LOGOUT_POLL_SECONDS" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid session") {
		t.Fatalf("expected session error, got %v", err)
	}
}

func TestLoadFromEnvInvalidAdminLockSeconds(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "ADMIN_LOCK_SECONDS" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid game timing") {
		t.Fatalf("expected game timing error, got %v", err)
	}
}

func TestLoadFromEnvProductionWithRedisAddr(t *testing.T) {
	cfg, err := LoadFromEnv(func(key string) string {
		switch key {
		case "APP_ENV":
			return "production"
		case "DATABASE_URL":
			return "postgres://localhost/numduel"
		case "JWT_SECRET":
			return "abcdefghijklmnopqrstuvwxyz123456"
		case "GAME_SECRET_PEPPER":
			return "abcdefghijklmnopqrstuvwxyz1234567890abcd"
		case "REDIS_ADDR":
			return "localhost:6379"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if !cfg.Production {
		t.Fatal("expected production")
	}
}

func TestLoadSuccess(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/numduel")
	t.Setenv("JWT_SECRET", "abcdefghijklmnopqrstuvwxyz123456")
	t.Setenv("GAME_SECRET_PEPPER", "abcdefghijklmnopqrstuvwxyz1234567890abcd")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		t.Fatalf("cfg: %+v", cfg)
	}
}

func TestEnvBoolParseTrueAndInvalid(t *testing.T) {
	if !envBool(func(string) string { return "true" }, "X", false) {
		t.Fatal("expected true")
	}
	if envBool(func(string) string { return "not-a-bool" }, "X", false) {
		t.Fatal("invalid bool should return fallback false")
	}
	if !envBool(func(string) string { return "not-a-bool" }, "X", true) {
		t.Fatal("invalid bool should return fallback true")
	}
}

func TestLoadFromEnvMasterCredentials(t *testing.T) {
	getenv := validGetenv()
	cfg, err := LoadFromEnv(func(key string) string {
		switch key {
		case "NUMDUEL_MASTER_EMAIL":
			return "admin@example.com"
		case "NUMDUEL_MASTER_PASSWORD":
			return "secretpass123"
		default:
			return getenv(key)
		}
	})
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.MasterEmail != "admin@example.com" || cfg.MasterPassword != "secretpass123" {
		t.Fatalf("master creds: %+v", cfg)
	}
}

func TestLoadFromEnvDefaultCronValues(t *testing.T) {
	cfg, err := LoadFromEnv(validGetenv())
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.BackupCron != "0 3 * * *" || cfg.RankingRebuildCron != "*/10 * * * *" {
		t.Fatalf("cron defaults: %+v", cfg)
	}
	if cfg.LogRetentionCron != "30 3 * * 0" || cfg.RefreshTokenCleanupCron != "0 4 * * *" {
		t.Fatalf("retention cron: %+v", cfg)
	}
	if cfg.RefreshTokenCleanupGraceDays != 7 || cfg.RetentionBatchSize != 1000 {
		t.Fatalf("retention defaults: %+v", cfg)
	}
}
