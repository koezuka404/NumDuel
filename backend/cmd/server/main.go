// アプリケーションのエントリポイント。
// 設定読み込み → DB 初期化 → 依存関係の組み立て → HTTP サーバー起動 の流れ。
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
	"github.com/numduel/numduel/internal/infrastructure/redis"
	"github.com/numduel/numduel/internal/router"
	"github.com/numduel/numduel/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	// 接続・マイグレーション・master シード・Repository 生成
	dbSetup, err := postgres.Setup(ctx, postgres.SetupConfig{
		DatabaseURL:       cfg.DatabaseURL,
		BackupDatabaseURL: cfg.BackupDatabaseURL,
		MasterEmail:       cfg.MasterEmail,
		MasterPassword:    cfg.MasterPassword,
	})
	if err != nil {
		log.Fatalf("database setup: %v", err)
	}

	jwtService, err := infrcrypto.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryMinutes)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}
	sessionStore := redis.NewStore() // 現状 no-op。Redis 実装後に差し替え
	authDeps := usecase.AuthDeps{
		Repo:                   dbSetup.Repo,
		Passwords:              infrcrypto.NewPasswordService(),
		AccessTokens:           jwtService,
		RefreshTokens:          infrcrypto.NewRefreshTokenService(),
		JWTRevoker:             sessionStore,
		WSSessions:             sessionStore,
		RefreshTokenExpiryDays: cfg.RefreshTokenExpiryDays,
	}

	e := echo.New()
	// ヘルスチェック（DB 接続確認のみ）
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
	router.Register(e, router.Deps{Auth: authDeps, JWT: jwtService, Cfg: cfg})

	go func() {
		addr := ":" + strconv.Itoa(cfg.Port)
		log.Printf("listening on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// SIGINT/SIGTERM で graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
