package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/router"
	"github.com/numduel/numduel/testutil"
	infrws "github.com/numduel/numduel/websocket"
	"github.com/numduel/numduel/usecase"
)

func TestRegisterWithoutWebSocket(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	authUC := testutil.NewAuthUC(t, repos)
	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}
	e := echo.New()
	router.Register(e, router.Deps{
		Auth:     authUC,
		Profile:  usecase.NewProfileUseCase(repos),
		Matching: testutil.NewMatchingUC(repos),
		Game:     testutil.NewGameUC(t, repos),
		Ranking:  testutil.NewRankingUC(repos),
		Admin:    testutil.NewAdminUC(repos, testutil.NewRankingUC(repos)),
		Cfg:      cfg,
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ws", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRegisterWithWebSocket(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	authUC := testutil.NewAuthUC(t, repos)
	hub := infrws.NewHub()
	gameUC := testutil.NewGameUCWithNotifier(t, repos, hub)
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, hub)
	wsHandler := &infrws.Handler{
		Hub: hub, WSAuth: wsAuth, Game: gameUC, JWTMin: 60,
	}
	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}
	e := echo.New()
	router.Register(e, router.Deps{
		Auth:     authUC,
		Profile:  usecase.NewProfileUseCase(repos),
		Matching: testutil.NewMatchingUC(repos),
		Game:     gameUC,
		Ranking:  testutil.NewRankingUC(repos),
		Admin:    testutil.NewAdminUC(repos, testutil.NewRankingUC(repos)),
		WSAuth:   wsAuth,
		WS:       wsHandler,
		JWT:      jwtSvc,
		AuthMW:   middleware.AuthConfig{JWT: jwtSvc, Repo: repos},
		Activity: middleware.ActivityUpdateConfig{Repo: repos},
		Cfg:      cfg,
	})

	found := false
	for _, r := range e.Routes() {
		if r.Method == http.MethodGet && r.Path == "/ws" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("/ws route not registered")
	}
}
