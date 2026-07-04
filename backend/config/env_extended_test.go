package config

import (
	"strings"
	"testing"
	"time"
)

func validGetenv() func(string) string {
	return func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgres://localhost/numduel"
		case "JWT_SECRET":
			return "abcdefghijklmnopqrstuvwxyz123456"
		case "GAME_SECRET_PEPPER":
			return "abcdefghijklmnopqrstuvwxyz1234567890abcd"
		default:
			return ""
		}
	}
}

func TestLoadFromEnvJWTSecretTooShort(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "JWT_SECRET" {
			return "short"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected JWT_SECRET error, got %v", err)
	}
}

func TestLoadFromEnvPepperTooShort(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "GAME_SECRET_PEPPER" {
			return "short"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "GAME_SECRET_PEPPER") {
		t.Fatalf("expected pepper error, got %v", err)
	}
}

func TestLoadFromEnvInvalidNumeric(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "PORT" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid numeric") {
		t.Fatalf("expected numeric error, got %v", err)
	}
}

func TestLoadFromEnvInvalidGameTiming(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "GAME_LOCK_SECONDS" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid game timing") {
		t.Fatalf("expected game timing error, got %v", err)
	}
}

func TestLoadFromEnvInvalidSession(t *testing.T) {
	getenv := validGetenv()
	_, err := LoadFromEnv(func(key string) string {
		if key == "SESSION_TIMEOUT_MINUTES" {
			return "0"
		}
		return getenv(key)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid session") {
		t.Fatalf("expected session error, got %v", err)
	}
}

func TestLoadFromEnvCORSAndWSOrigins(t *testing.T) {
	getenv := validGetenv()
	cfg, err := LoadFromEnv(func(key string) string {
		switch key {
		case "CORS_ALLOWED_ORIGINS":
			return " http://a.test , ,http://b.test "
		case "WS_ALLOWED_ORIGINS":
			return "ws://a.test"
		default:
			return getenv(key)
		}
	})
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || len(cfg.WSAllowedOrigins) != 1 {
		t.Fatalf("origins: cors=%v ws=%v", cfg.CORSAllowedOrigins, cfg.WSAllowedOrigins)
	}
}

func TestLoadFromEnvCustomValues(t *testing.T) {
	getenv := validGetenv()
	cfg, err := LoadFromEnv(func(key string) string {
		switch key {
		case "PORT", "JWT_EXPIRY_MINUTES", "REFRESH_TOKEN_EXPIRY_DAYS":
			return "10"
		case "COOKIE_SECURE":
			return "true"
		case "REDIS_ADDR":
			return "localhost:6379"
		case "ENV":
			return "prod"
		default:
			return getenv(key)
		}
	})
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.Port != 10 || !cfg.CookieSecure || !cfg.Production {
		t.Fatalf("cfg: %+v", cfg)
	}
}

func TestConfigDurationHelpers(t *testing.T) {
	cfg := &Config{
		GameLockSeconds: 2, AdminLockSeconds: 5, TurnDurationSeconds: 30,
		TurnTimeoutPollSeconds: 1, SecretSetupSeconds: 60, SecretTimeoutPollSeconds: 1,
		SessionTimeoutMinutes: 5, AutoLogoutPollSeconds: 60, RetentionBatchSleepMs: 100,
	}
	if cfg.GameLockTTL() != 2*time.Second || cfg.AdminLockTTL() != 5*time.Second {
		t.Fatal("lock ttl")
	}
	if cfg.TurnDuration() != 30*time.Second || cfg.SecretSetupDuration() != 60*time.Second {
		t.Fatal("game duration")
	}
	if cfg.SessionTimeout() != 5*time.Minute || cfg.AutoLogoutPollInterval() != 60*time.Second {
		t.Fatal("session")
	}
	if cfg.RetentionBatchSleep() != 100*time.Millisecond {
		t.Fatal("retention sleep")
	}
	if cfg.TurnTimeoutPollInterval() != time.Second || cfg.SecretTimeoutPollInterval() != time.Second {
		t.Fatal("poll intervals")
	}
}

func TestIsProductionEnv(t *testing.T) {
	if !isProductionEnv("production", "") || !isProductionEnv("", "prod") {
		t.Fatal("production env")
	}
	if isProductionEnv("development", "") {
		t.Fatal("non production")
	}
}

func TestRedisAddrConfigured(t *testing.T) {
	if redisAddrConfigured(func(string) string { return "" }) != "" {
		t.Fatal("empty redis")
	}
	if redisAddrConfigured(func(k string) string {
		if k == "REDIS_URL" {
			return "redis://localhost"
		}
		return ""
	}) == "" {
		t.Fatal("redis url")
	}
	if redisAddrConfigured(func(k string) string {
		if k == "REDIS_ADDR" {
			return "localhost:6379"
		}
		return ""
	}) == "" {
		t.Fatal("redis addr")
	}
}

func TestEnvHelpersFallback(t *testing.T) {
	getenv := func(string) string { return "" }
	if envInt(getenv, "X", 7) != 7 {
		t.Fatal("envInt fallback")
	}
	if envInt(func(string) string { return "bad" }, "X", 7) != 7 {
		t.Fatal("envInt parse error")
	}
	if envBool(getenv, "X", true) != true {
		t.Fatal("envBool fallback")
	}
	if envBool(func(string) string { return "false" }, "X", true) {
		t.Fatal("envBool false")
	}
	if envString(getenv, "X", "fb") != "fb" {
		t.Fatal("envString fallback")
	}
	if envString(func(string) string { return "custom" }, "X", "fb") != "custom" {
		t.Fatal("envString value")
	}
}

func TestLoadCallsLoadFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected Load error without env")
	}
}
