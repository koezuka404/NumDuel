package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	"github.com/numduel/numduel/controller"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

type ctrlEnv struct {
	echo  *echo.Echo
	repos repository.Repos
}

func setupCtrlEnv(t *testing.T) *ctrlEnv {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	authUC := testutil.NewAuthUC(t, repos)
	profileUC := usecase.NewProfileUseCase(repos)
	gameUC := testutil.NewGameUC(t, repos)
	matchingUC := testutil.NewMatchingUC(repos)
	rankingUC := testutil.NewRankingUC(repos)
	adminUC := testutil.NewAdminUC(repos, rankingUC)
	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}

	auth := controller.NewAuthController(authUC, cfg.CookieSecure, cfg.JWTExpiryMinutes, cfg.RefreshTokenExpiryDays)
	me := controller.NewMeController(authUC, profileUC)
	match := controller.NewMatchingController(matchingUC)
	game := controller.NewGameController(gameUC)
	ranking := controller.NewRankingController(rankingUC)
	admin := controller.NewAdminController(adminUC)

	e := echo.New()
	api := e.Group("/api")
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh)

	protected := api.Group("", middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Repo: repos}))
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
	adminGroup.GET("/logs/types", admin.ListLogTypes)
	adminGroup.GET("/logs", admin.SearchLogs)
	adminGroup.GET("/logs/download", admin.DownloadLogs)
	adminGroup.GET("/backup/status", admin.BackupStatus)

	return &ctrlEnv{echo: e, repos: repos}
}

func (env *ctrlEnv) do(t *testing.T, method, path string, cookies []*http.Cookie, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		payload = bytes.NewReader(b)
	} else {
		payload = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, payload)
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	env.echo.ServeHTTP(rec, req)
	return rec
}

func (env *ctrlEnv) login(t *testing.T, email, password string) []*http.Cookie {
	t.Helper()
	auth := testutil.NewAuthUC(t, env.repos)
	out, err := auth.Login(t.Context(), usecase.LoginInput{Email: email, Password: password})
	if err != nil {
		t.Fatalf("login %s: %v", email, err)
	}
	return []*http.Cookie{
		{Name: middleware.AccessCookieName, Value: out.AccessToken},
		{Name: middleware.RefreshCookieName, Value: out.RefreshToken, Path: "/api/auth/refresh"},
	}
}

func (env *ctrlEnv) seedHistories(t *testing.T, userID uuid.UUID) {
	t.Helper()
	ctx := t.Context()
	now := time.Now().UTC()
	opponent := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	gameID := uuid.New()
	history := &model.MatchHistory{
		ID: uuid.New(), GameID: gameID, WinnerID: userID, LoserID: opponent.ID,
		WinnerUsername: "alice", LoserUsername: "bob", FinishedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := env.repos.MatchHistory.Create(ctx, history); err != nil {
		t.Fatalf("match history: %v", err)
	}
	loginLog := &model.LoginLog{
		ID: uuid.New(), UserID: userID, Action: model.LoginActionLogin, CreatedAt: now, UpdatedAt: now,
	}
	if err := env.repos.LoginLog.Create(ctx, loginLog); err != nil {
		t.Fatalf("login log: %v", err)
	}
	wsLog := &model.WSConnectionLog{
		ID: uuid.New(), UserID: userID, ConnectionID: "c1", ConnectedAt: now,
	}
	if err := env.repos.WSConnectionLog.Create(ctx, wsLog); err != nil {
		t.Fatalf("ws log: %v", err)
	}
}
