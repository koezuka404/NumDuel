// HTTP ルート定義。Middleware → Controller への振り分けのみ行う。
package router

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/config"
	"github.com/numduel/numduel/internal/controller"
	infrcrypto "github.com/numduel/numduel/internal/infrastructure/crypto"
	"github.com/numduel/numduel/internal/middleware"
	"github.com/numduel/numduel/internal/usecase"
)

// Deps はルーター登録に必要な依存関係。
type Deps struct {
	Auth usecase.AuthDeps
	JWT  *infrcrypto.JWTService
	Cfg  *config.Config
}

// Register は /api 配下の認証ルートを Echo に登録する。
func Register(e *echo.Echo, deps Deps) {
	auth := controller.NewAuthController(deps.Auth, deps.Cfg.CookieSecure, deps.Cfg.RefreshTokenExpiryDays)
	me := controller.NewMeController(deps.Auth)

	api := e.Group("/api")
	// 認証不要
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh) // Cookie の refresh_token を使用

	// JWT 必須
	protected := api.Group("", middleware.Auth(deps.JWT, deps.Auth.JWTRevoker))
	protected.POST("/auth/logout", auth.Logout)
	protected.GET("/me", me.Get)
}
