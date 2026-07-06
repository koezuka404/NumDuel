package router

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	"github.com/numduel/numduel/controller"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
	infrws "github.com/numduel/numduel/websocket"
)

type Deps struct {
	Auth     usecase.IAuthUsecase
	Profile  usecase.IProfileUsecase
	Matching usecase.IMatchingUsecase
	Game     usecase.IGameUsecase
	Ranking  usecase.IRankingUsecase
	Admin    usecase.IAdminUsecase
	WSAuth   usecase.IWSAuthUsecase
	WS       *infrws.Handler
	JWT      *infrcrypto.JWTService
	AuthMW   middleware.AuthConfig
	Activity middleware.ActivityUpdateConfig
	Cfg      *config.Config
}

func Register(e *echo.Echo, deps Deps) {
	auth := controller.NewAuthController(deps.Auth, deps.Cfg.CookieSecure, deps.Cfg.JWTExpiryMinutes, deps.Cfg.RefreshTokenExpiryDays)
	me := controller.NewMeController(deps.Auth, deps.Profile)
	match := controller.NewMatchingController(deps.Matching)
	game := controller.NewGameController(deps.Game)
	ranking := controller.NewRankingController(deps.Ranking)
	admin := controller.NewAdminController(deps.Admin)
	wsAuthCtrl := controller.NewWSAuthController(deps.WSAuth)

	api := e.Group("/api", middleware.RateLimitPublic())
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh)
	api.GET("/auth/session", auth.Session, middleware.TryAuth(deps.AuthMW))

	protected := api.Group("",
		middleware.Auth(deps.AuthMW),
		middleware.UserRateLimit(),
		middleware.ActivityUpdate(deps.Activity),
	)
	protected.POST("/auth/logout", auth.Logout)
	protected.GET("/auth/ws-ticket", wsAuthCtrl.IssueTicket)
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
	adminGroup.GET("/logs/types", admin.ListLogTypes)
	adminGroup.GET("/logs", admin.SearchLogs)
	adminGroup.GET("/logs/download", admin.DownloadLogs)
	adminGroup.GET("/backup/status", admin.BackupStatus)

	if deps.WS != nil {
		e.GET("/ws", deps.WS.Handle)
	}
}
