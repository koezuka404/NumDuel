// 全 API リクエストを activity_logs に記録（パスワード・JWT 本文は含めない）
package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// RequestLogConfig は RequestLog の依存関係
type RequestLogConfig struct {
	Repo repository.Repos
}

// RequestLog は HTTP メタデータを activity_logs（log_type: http_request）へ非同期 INSERT する
// /health と OPTIONS は記録しない。login/register 等もリクエスト本文は保存しない
func RequestLog(cfg RequestLogConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if shouldSkipRequestLog(c) {
				return err
			}
			if cfg.Repo == nil {
				return err
			}
			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}
			var userID *uuid.UUID
			if auth, ok := AuthFrom(c); ok {
				uid := auth.UserID
				userID = &uid
			}
			detail, marshalErr := json.Marshal(map[string]any{
				"method":     c.Request().Method,
				"path":       c.Path(),
				"status":     status,
				"durationMs": time.Since(start).Milliseconds(),
				"ip":         c.RealIP(),
			})
			if marshalErr != nil {
				log.Printf("request log: marshal detail: %v", marshalErr)
				return err
			}
			now := time.Now().UTC()
			entry := &model.ActivityLog{
				ID: uuid.New(), UserID: userID, LogType: "http_request",
				Detail: detail, CreatedAt: now, UpdatedAt: now,
			}
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				if createErr := cfg.Repo.ActivityLog.Create(ctx, entry); createErr != nil {
					log.Printf("request log: create activity log: %v", createErr)
				}
			}()
			return err
		}
	}
}

func shouldSkipRequestLog(c echo.Context) bool {
	if c.Request().Method == http.MethodOptions {
		return true
	}
	path := c.Path()
	if path == "/health" {
		return true
	}
	// ログ閲覧 API 自体は activity_logs を汚さない
	return strings.HasPrefix(path, "/api/admin/logs")
}
