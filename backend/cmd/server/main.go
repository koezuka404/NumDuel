// アプリケーションのエントリポイント。
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/config"
	infrcrypto "github.com/numduel/numduel/internal/infrastructure/crypto"
	"github.com/numduel/numduel/internal/infrastructure/postgres"
	infrredis "github.com/numduel/numduel/internal/infrastructure/redis"
	infrws "github.com/numduel/numduel/internal/infrastructure/websocket"
	"github.com/numduel/numduel/internal/middleware"
	"github.com/numduel/numduel/internal/router"
	"github.com/numduel/numduel/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	dbSetup, err := postgres.Setup(ctx, postgres.SetupConfig{
		DatabaseURL:       cfg.DatabaseURL,
		BackupDatabaseURL: cfg.BackupDatabaseURL,
		MasterEmail:       cfg.MasterEmail,
		MasterPassword:    cfg.MasterPassword,
	})
	if err != nil {
		log.Fatalf("database setup: %v", err)
	}

	redisStore, err := infrredis.Open(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func() { _ = redisStore.Close() }()

	jwtService, err := infrcrypto.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryMinutes)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}
	secretHasher, err := infrcrypto.NewSecretHashService(cfg.GameSecretPepper)
	if err != nil {
		log.Fatalf("secret hasher: %v", err)
	}

	hub := infrws.NewHub()
	sessionStore := infrws.NewSessionStore(hub, redisStore)

	authDeps := usecase.AuthDeps{
		Repo:                   dbSetup.Repo,
		Passwords:              infrcrypto.NewPasswordService(),
		AccessTokens:           jwtService,
		RefreshTokens:          infrcrypto.NewRefreshTokenService(),
		JWTRevoker:             redisStore,
		WSSessions:             sessionStore,
		RefreshTokenExpiryDays: cfg.RefreshTokenExpiryDays,
	}
	gameDeps := usecase.GameDeps{
		Repo: dbSetup.Repo, Secrets: secretHasher,
		Locks: redisStore, Turns: redisStore, Notifier: hub,
		TurnDuration: cfg.TurnDuration(), GameLockTTL: cfg.GameLockTTL(),
	}
	matchingDeps := usecase.MatchingDeps{Repo: dbSetup.Repo, Notifier: hub}
	wsAuthDeps := usecase.WSAuthDeps{
		Repo: dbSetup.Repo, JWT: jwtService,
		Revoker: redisStore, ForceLogout: redisStore, Notifier: hub,
	}

	allowed := make(map[string]struct{}, len(cfg.WSAllowedOrigins))
	for _, o := range cfg.WSAllowedOrigins {
		allowed[o] = struct{}{}
	}
	wsHandler := &infrws.Handler{
		Hub: hub, WSAuth: wsAuthDeps, Game: gameDeps,
		Allowed: allowed, Redis: redisStore,
		JWTMin: cfg.JWTExpiryMinutes, Repo: dbSetup.Repo,
	}

	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		pingCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := dbSetup.Primary.Ping(pingCtx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{"code": "internal_error", "message": "database unavailable"},
			})
		}
		return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
	})

	router.Register(e, router.Deps{
		Auth: authDeps, Matching: matchingDeps, Game: gameDeps,
		WSAuth: wsAuthDeps, WS: wsHandler, JWT: jwtService,
		AuthMW: middleware.AuthConfig{
			JWT: jwtService, Revoker: redisStore,
			ForceLogout: redisStore, Repo: dbSetup.Repo,
		},
		Cfg: cfg,
	})

	go func() {
		addr := ":" + strconv.Itoa(cfg.Port)
		log.Printf("listening on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
