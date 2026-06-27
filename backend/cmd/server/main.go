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
	"github.com/numduel/numduel/internal/infrastructure/postgres"
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
	_ = dbSetup.Repo
	_ = dbSetup.Syncer

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
