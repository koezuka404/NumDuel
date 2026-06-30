// HTTP ルート定義Middleware → Controller への振り分けのみ行う
package router

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	"github.com/numduel/numduel/controller"
	infrcrypto "github.com/numduel/numduel/crypto"
	infrws "github.com/numduel/numduel/websocket"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type Deps struct {
	Auth     usecase.AuthDeps
	Profile  usecase.ProfileDeps
	Matching usecase.MatchingDeps
	Game     usecase.GameDeps
	Ranking  usecase.RankingDeps
	Admin    usecase.AdminDeps
	WSAuth   usecase.WSAuthDeps
	WS       *infrws.Handler
	JWT      *infrcrypto.JWTService
	AuthMW   middleware.AuthConfig
	Activity middleware.ActivityUpdateConfig // JWT 必須ルートで last_activity_at 更新
	Cfg      *config.Config
}

func Register(e *echo.Echo, deps Deps) {
	auth := controller.NewAuthController(deps.Auth, deps.Cfg.CookieSecure, deps.Cfg.JWTExpiryMinutes, deps.Cfg.RefreshTokenExpiryDays)
	me := controller.NewMeController(deps.Auth, deps.Profile)
	match := controller.NewMatchingController(deps.Matching)
	game := controller.NewGameController(deps.Game)
	ranking := controller.NewRankingController(deps.Ranking)
	admin := controller.NewAdminController(deps.Admin)

	api := e.Group("/api", middleware.RateLimit())
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh)

	// Auth 通過後: last_activity_at 更新 → AutoLogout 判定の延長
	protected := api.Group("",
		middleware.Auth(deps.AuthMW),
		middleware.ActivityUpdate(deps.Activity),
	)
	protected.POST("/auth/logout", auth.Logout)
	protected.GET("/me", me.Get)
	protected.GET("/me/profile", me.GetProfile)
	protected.GET("/me/match-history", me.MatchHistory)
	protected.GET("/me/login-history", me.LoginHistory)
	protected.GET("/me/ws-history", me.WSHistory)
	protected.POST("/matching/start", match.Start)
	protected.POST("/matching/cancel", match.Cancel)
	protected.GET("/matching/status", match.Status)
	protected.GET("/games/:id", game.Get)
	protected.GET("/ranking", ranking.Get)

	adminGroup := protected.Group("/admin", middleware.Admin())
	adminGroup.GET("/users", admin.ListUsers)
	adminGroup.GET("/users/search", admin.SearchUsers)
	adminGroup.DELETE("/users/:id", admin.DeleteUser)
	adminGroup.POST("/ranking/rebuild", admin.RebuildRanking)
	adminGroup.GET("/logs", admin.SearchLogs)
	adminGroup.GET("/logs/download", admin.DownloadLogs)
	adminGroup.GET("/backup/status", admin.BackupStatus)

	if deps.WS != nil {
		e.GET("/ws", deps.WS.Handle)
	}
}
