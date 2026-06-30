// アプリケーションのエントリポイント
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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/numduel/numduel/config"
	"github.com/numduel/numduel/db"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	infrredis "github.com/numduel/numduel/redis"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/router"
	"github.com/numduel/numduel/usecase"
	infrws "github.com/numduel/numduel/websocket"
	"github.com/numduel/numduel/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dbSetup, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL:       cfg.DatabaseURL,
		BackupDatabaseURL: cfg.BackupDatabaseURL,
		Migrate:           os.Getenv("SKIP_MIGRATE") != "1",
	})
	if err != nil {
		log.Fatalf("database setup: %v", err)
	}
	defer closeGormDB(dbSetup.Primary.Gorm())
	if dbSetup.Backup != nil {
		defer closeGormDB(dbSetup.Backup.Gorm())
	}

	jwtService, err := infrcrypto.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryMinutes)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}
	secretHasher, err := infrcrypto.NewSecretHashService(cfg.GameSecretPepper)
	if err != nil {
		log.Fatalf("secret hasher: %v", err)
	}

	rdb := db.OpenRedis()
	if rdb != nil {
		defer rdb.Close()
	}
	redisStore := infrredis.NewStore(rdb)

	hub := infrws.NewHub()
	sessionStore := infrws.NewSessionStore(hub, redisStore)

	authDeps := usecase.AuthDeps{
		Repo:                   dbSetup.Repo,
		Tx:                     dbSetup.Tx,
		Passwords:              infrcrypto.NewPasswordService(),
		AccessTokens:           jwtService,
		RefreshTokens:          infrcrypto.NewRefreshTokenService(),
		JWTRevoker:             redisStore,
		WSSessions:             sessionStore,
		RefreshTokenExpiryDays: cfg.RefreshTokenExpiryDays,
	}
	gameDeps := usecase.GameDeps{
		Repo: dbSetup.Repo, Tx: dbSetup.Tx, Secrets: secretHasher,
		Locks: redisStore, Turns: redisStore, Random: infrcrypto.NewRandomNumberService(), Notifier: hub,
		TurnDuration: cfg.TurnDuration(), SecretSetup: cfg.SecretSetupDuration(),
		GameLockTTL: cfg.GameLockTTL(),
	}
	matchingDeps := usecase.MatchingDeps{Repo: dbSetup.Repo, Tx: dbSetup.Tx, Notifier: hub}
	profileDeps := usecase.ProfileDeps{Repo: dbSetup.Repo}
	rankingDeps := usecase.RankingDeps{Repo: dbSetup.Repo, Tx: dbSetup.Tx}
	adminDeps := usecase.AdminDeps{
		Repo: dbSetup.Repo, Tx: dbSetup.Tx, WSSessions: sessionStore, BackupStatus: redisStore,
	}
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

	autoLogoutDeps := usecase.AutoLogoutDeps{
		Repo: dbSetup.Repo, Tx: dbSetup.Tx, ForceLogout: redisStore,
		SessionTimeout: cfg.SessionTimeout(),
		// WS 接続中なら unauthorized を返してから切断
		ForceDisconnect: func(ctx context.Context, userID uuid.UUID) error {
			return sessionStore.DisconnectWithError(ctx, userID, model.CodeUnauthorized, "invalid credentials")
		},
	}

	if err := usecase.RecoverActiveGames(context.Background(), gameDeps); err != nil {
		log.Printf("recover active games: %v", err)
	}

	e := echo.New()
	e.Use(
		middleware.Recover(),
		middleware.CORS(cfg.CORSAllowedOrigins),
	)
	e.GET("/health", func(c echo.Context) error {
		pingCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(pingCtx, dbSetup.Primary.Gorm()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{"code": "internal_error", "message": "database unavailable"},
			})
		}
		return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
	})

	router.Register(e, router.Deps{
		Auth: authDeps, Profile: profileDeps, Matching: matchingDeps, Game: gameDeps,
		Ranking: rankingDeps, Admin: adminDeps,
		WSAuth: wsAuthDeps, WS: wsHandler, JWT: jwtService,
		AuthMW: middleware.AuthConfig{
			JWT: jwtService, Revoker: redisStore,
			ForceLogout: redisStore, Repo: dbSetup.Repo,
		},
		// protected API ごとに last_activity_at を更新（AutoLogoutWorker 連携）
		Activity: middleware.ActivityUpdateConfig{Repo: dbSetup.Repo},
		Cfg: cfg,
	})

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	if redisStore != nil && cfg.TurnTimeoutPollSeconds > 0 {
		go (&worker.TurnTimeoutWorker{
			Store:    redisStore,
			Game:     gameDeps,
			Interval: cfg.TurnTimeoutPollInterval(),
		}).Run(workerCtx)
	}
	if cfg.SecretTimeoutPollSeconds > 0 {
		go (&worker.SecretSetupTimeoutWorker{
			Game:     gameDeps,
			Interval: cfg.SecretTimeoutPollInterval(),
		}).Run(workerCtx)
	}
	if redisStore != nil && cfg.AutoLogoutPollSeconds > 0 {
		// last_activity_at が SESSION_TIMEOUT_MINUTES を超えたユーザーを自動ログアウト
		go (&worker.AutoLogoutWorker{
			Deps:     autoLogoutDeps,
			Interval: cfg.AutoLogoutPollInterval(),
		}).Run(workerCtx)
	}

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

func closeGormDB(gdb *gorm.DB) {
	sqlDB, err := db.SQLDB(gdb)
	if err != nil {
		log.Printf("database sql handle: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("database close: %v", err)
	}
}
