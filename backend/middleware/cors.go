// CORS_ALLOWED_ORIGINS で許可するオリジンのみ API へ credentials 付きアクセスを許可
package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS はフロントからの Cookie 付き API 呼び出し向け CORS ヘッダを付与する
// origins 空なら no-op（ローカル直叩き用。WS は WS_ALLOWED_ORIGINS で別管理）
func CORS(origins []string) echo.MiddlewareFunc {
	if len(origins) == 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderCookie,
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie"},
	})
}
