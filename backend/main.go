//アプリケーションのエントリポイント
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
	defer closeGormDB(dbSetup.Primary)
	if dbSetup.Backup != nil {
		defer closeGormDB(dbSetup.Backup)
	}

	jwtService, err := infrcrypto.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryMinutes)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}
	secretHasher, err := infrcrypto.NewSecretHashService(cfg.GameSecretPepper)
	if err != nil {
		log.Fatalf("secret hasher: %v", err)
	}

	rdb, err := db.OpenRedis(cfg.Production)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	if rdb != nil {
		defer rdb.Close()
	}
	redisStore := infrredis.NewStore(rdb)

	hub := infrws.NewHub()
	sessionStore := infrws.NewSessionStore(hub, redisStore)
	refreshGen := infrcrypto.NewRefreshTokenService()

	authUC := usecase.NewAuthUseCase(
		dbSetup.Repos,
		infrcrypto.NewPasswordService(),
		jwtService,
		refreshGen,
		redisStore,
		sessionStore,
		cfg.RefreshTokenExpiryDays,
		cfg.RefreshTokenCleanupGraceDays,
	)
	gameUC := usecase.NewGameUseCase(
		dbSetup.Repos, secretHasher, redisStore, redisStore,
		infrcrypto.NewRandomNumberService(), hub,
		cfg.TurnDuration(), cfg.SecretSetupDuration(), cfg.GameLockTTL(),
	)
	matchingUC := usecase.NewMatchingUseCase(dbSetup.Repos, hub)
	profileUC := usecase.NewProfileUseCase(dbSetup.Repos)
	rankingUC := usecase.NewRankingUseCase(dbSetup.Repos, redisStore, cfg.AdminLockTTL())
	adminUC := usecase.NewAdminUseCase(dbSetup.Repos, rankingUC, sessionStore, redisStore, redisStore, redisStore, cfg.AdminLockTTL())
	wsAuthUC := usecase.NewWSAuthUseCase(dbSetup.Repos, jwtService, redisStore, redisStore, hub)
	autoLogoutUC := usecase.NewAutoLogoutUseCase(dbSetup.Repos, redisStore, func(ctx context.Context, userID uuid.UUID) error {
		return sessionStore.DisconnectWithError(ctx, userID, "unauthorized", "invalid credentials")
	}, cfg.SessionTimeout())
	backupUC := usecase.NewBackupUseCase(dbSetup.Syncer, redisStore, 0)
	logRetentionUC := usecase.NewLogRetentionUseCase(
		dbSetup.Repos,
		cfg.ActivityLogRetentionDays,
		cfg.LoginLogRetentionDays,
		cfg.WSLogRetentionDays,
		cfg.RetentionBatchSize,
		cfg.RetentionBatchSleep(),
	)

	allowed := make(map[string]struct{}, len(cfg.WSAllowedOrigins))
	for _, o := range cfg.WSAllowedOrigins {
		allowed[o] = struct{}{}
	}
	wsHandler := &infrws.Handler{
		Hub: hub, WSAuth: wsAuthUC, Game: gameUC,
		Allowed: allowed, Redis: redisStore,
		JWTMin: cfg.JWTExpiryMinutes,
	}

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		log.Printf("recover active games: %v", err)
	}
	if err := authUC.SeedMaster(context.Background(), usecase.SeedMasterInput{
		Email: cfg.MasterEmail, Password: cfg.MasterPassword,
	}); err != nil {
		log.Printf("seed master: %v", err)
	}

	e := echo.New()
	//Recover→CORS→RequestLog（global）/api配下はRateLimit→Auth→ActivityUpdate→Admin
	e.Use(
		middleware.Recover(),
		middleware.CORS(cfg.CORSAllowedOrigins),
		middleware.RequestLog(middleware.RequestLogConfig{Repo: dbSetup.Repos}),
	)
	e.GET("/health", func(c echo.Context) error {
		pingCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(pingCtx, dbSetup.Primary); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]any{
				"error": map[string]string{"code": "internal_error", "message": "database unavailable"},
			})
		}
		return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
	})

	router.Register(e, router.Deps{
		Auth: authUC, Profile: profileUC, Matching: matchingUC, Game: gameUC,
		Ranking: rankingUC, Admin: adminUC,
		WSAuth: wsAuthUC, WS: wsHandler, JWT: jwtService,
		AuthMW: middleware.AuthConfig{
			JWT: jwtService, Revoker: redisStore,
			ForceLogout: redisStore, Repo: dbSetup.Repos,
		},
		Activity: middleware.ActivityUpdateConfig{Repo: dbSetup.Repos},
		Cfg: cfg,
	})

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	if redisStore != nil && cfg.TurnTimeoutPollSeconds > 0 {
		go (&worker.TurnTimeoutWorker{
			Store:    redisStore,
			Game:     gameUC,
			Interval: cfg.TurnTimeoutPollInterval(),
		}).Run(workerCtx)
	}
	if cfg.SecretTimeoutPollSeconds > 0 {
		go (&worker.SecretSetupTimeoutWorker{
			Game:     gameUC,
			Interval: cfg.SecretTimeoutPollInterval(),
		}).Run(workerCtx)
	}
	if redisStore != nil && cfg.AutoLogoutPollSeconds > 0 {
		go (&worker.AutoLogoutWorker{
			AutoLogout: autoLogoutUC,
			Interval:   cfg.AutoLogoutPollInterval(),
		}).Run(workerCtx)
	}
	if dbSetup.Syncer != nil && redisStore != nil && cfg.BackupCron != "" {
		go (&worker.BackupWorker{
			Backup: backupUC,
			Cron:   cfg.BackupCron,
		}).Run(workerCtx)
	}
	if cfg.RankingRebuildCron != "" {
		go (&worker.RankingRebuildWorker{
			Ranking: rankingUC,
			Cron:    cfg.RankingRebuildCron,
		}).Run(workerCtx)
	}
	if cfg.LogRetentionCron != "" {
		go (&worker.LogRetentionWorker{
			Retention: logRetentionUC,
			Cron:      cfg.LogRetentionCron,
		}).Run(workerCtx)
	}
	if cfg.RefreshTokenCleanupCron != "" {
		go (&worker.RefreshTokenCleanupWorker{
			Auth: authUC,
			Cron: cfg.RefreshTokenCleanupCron,
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
