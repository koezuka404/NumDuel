package config

import (
	"strings"
	"testing"
)

func TestLoadFromEnvRequiresDatabaseURL(t *testing.T) {
	_, err := LoadFromEnv(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLoadFromEnvValidMinimal(t *testing.T) {
	getenv := func(key string) string {
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
	cfg, err := LoadFromEnv(getenv)
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.Port != 8090 || cfg.JWTExpiryMinutes != 60 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadFromEnvProductionRequiresRedis(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "APP_ENV":
			return "production"
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
	_, err := LoadFromEnv(getenv)
	if err == nil || !strings.Contains(err.Error(), "REDIS") {
		t.Fatalf("expected REDIS required error, got %v", err)
	}
}

func TestLoadFromEnvREDISURL(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "APP_ENV":
			return "production"
		case "DATABASE_URL":
			return "postgres://localhost/numduel"
		case "JWT_SECRET":
			return "abcdefghijklmnopqrstuvwxyz123456"
		case "GAME_SECRET_PEPPER":
			return "abcdefghijklmnopqrstuvwxyz1234567890abcd"
		case "REDIS_URL":
			return "redis://localhost:6379/0"
		default:
			return ""
		}
	}
	cfg, err := LoadFromEnv(getenv)
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if !cfg.Production {
		t.Fatal("expected production=true")
	}
}
