//CORS_ALLOWED_ORIGINSで許可するオリジンのみAPIへcredentials付きアクセスを許可
package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

//CORSはフロントからのCookie付きAPI呼び出し向けCORSヘッダを付与する
//origins空ならno-op（ローカル直叩き用。WSはWS_ALLOWED_ORIGINSで別管理）
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
