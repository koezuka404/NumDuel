// HTTP ルート定義。Middleware → Controller への振り分けのみ行う。
package router

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/config"
	"github.com/numduel/numduel/internal/controller"
	infrcrypto "github.com/numduel/numduel/internal/infrastructure/crypto"
	infrws "github.com/numduel/numduel/internal/infrastructure/websocket"
	"github.com/numduel/numduel/internal/middleware"
	"github.com/numduel/numduel/internal/usecase"
)

type Deps struct {
	Auth     usecase.AuthDeps
	Matching usecase.MatchingDeps
	Game     usecase.GameDeps
	WSAuth   usecase.WSAuthDeps
	WS       *infrws.Handler
	JWT      *infrcrypto.JWTService
	AuthMW   middleware.AuthConfig
	Cfg      *config.Config
}

func Register(e *echo.Echo, deps Deps) {
	auth := controller.NewAuthController(deps.Auth, deps.Cfg.CookieSecure, deps.Cfg.RefreshTokenExpiryDays)
	me := controller.NewMeController(deps.Auth)
	match := controller.NewMatchingController(deps.Matching)
	game := controller.NewGameController(deps.Game)

	api := e.Group("/api")
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh)

	protected := api.Group("", middleware.Auth(deps.AuthMW))
	protected.POST("/auth/logout", auth.Logout)
	protected.GET("/me", me.Get)
	protected.POST("/matching/start", match.Start)
	protected.POST("/matching/cancel", match.Cancel)
	protected.GET("/matching/status", match.Status)
	protected.GET("/games/:id", game.Get)

	if deps.WS != nil {
		e.GET("/ws", deps.WS.Handle)
	}
}
