package websocket_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	goredis "github.com/redis/go-redis/v9"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/config"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/repository"
	infrredis "github.com/numduel/numduel/redis"
	"github.com/numduel/numduel/router"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
	infrws "github.com/numduel/numduel/websocket"
)

type wsTestOpts struct {
	allowed      map[string]struct{}
	redis        usecase.IWSSessionStore
	recordConnErr error
}

type wsAuthStub struct {
	inner         *usecase.WSAuthUseCase
	recordConnErr error
}

func (s *wsAuthStub) Authenticate(ctx context.Context, token string) (*usecase.WSAuthOutput, error) {
	return s.inner.Authenticate(ctx, token)
}

func (s *wsAuthStub) AuthenticateByTicket(ctx context.Context, ticket string) (*usecase.WSAuthOutput, error) {
	return s.inner.AuthenticateByTicket(ctx, ticket)
}

func (s *wsAuthStub) IssueTicket(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.inner.IssueTicket(ctx, userID)
}

func (s *wsAuthStub) NotifyOpponentConnected(ctx context.Context, userID uuid.UUID) {
	s.inner.NotifyOpponentConnected(ctx, userID)
}

func (s *wsAuthStub) RecordConnection(ctx context.Context, userID uuid.UUID, connectionID string) (uuid.UUID, error) {
	if s.recordConnErr != nil {
		return uuid.Nil, s.recordConnErr
	}
	return s.inner.RecordConnection(ctx, userID, connectionID)
}

func (s *wsAuthStub) TouchActivity(ctx context.Context, userID uuid.UUID) {
	s.inner.TouchActivity(ctx, userID)
}

func (s *wsAuthStub) CloseConnectionLog(ctx context.Context, logID uuid.UUID) {
	s.inner.CloseConnectionLog(ctx, logID)
}

func (s *wsAuthStub) NotifyOpponentDisconnected(ctx context.Context, userID uuid.UUID) {
	s.inner.NotifyOpponentDisconnected(ctx, userID)
}

type testEnv struct {
	echo  *echo.Echo
	repos repository.Repos
	auth  *usecase.AuthUseCase
	hub   *infrws.Hub
	srv   *httptest.Server
}

func setupWSTest(t *testing.T) *testEnv {
	t.Helper()
	return setupWSTestWithOpts(t, wsTestOpts{})
}

func setupWSTestWithOpts(t *testing.T, opts wsTestOpts) *testEnv {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	authUC := testutil.NewAuthUC(t, repos)
	hub := infrws.NewHub()
	gameUC := testutil.NewGameUCWithNotifier(t, repos, hub)
	matchingUC := testutil.NewMatchingUC(repos)
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, hub, nil)
	var wsAuthUC usecase.IWSAuthUsecase = wsAuth
	if opts.recordConnErr != nil {
		wsAuthUC = &wsAuthStub{inner: wsAuth, recordConnErr: opts.recordConnErr}
	}
	wsHandler := &infrws.Handler{
		Hub: hub, WSAuth: wsAuthUC, Game: gameUC, JWTMin: 60,
		Allowed: opts.allowed,
		Redis:   opts.redis,
	}
	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}
	e := echo.New()
	router.Register(e, router.Deps{
		Auth: authUC,
		Profile: usecase.NewProfileUseCase(repos),
		Matching: matchingUC,
		Game: gameUC,
		Ranking: testutil.NewRankingUC(repos),
		Admin: testutil.NewAdminUC(repos, testutil.NewRankingUC(repos)),
		WSAuth: wsAuthUC,
		WS: wsHandler,
		JWT: jwtSvc,
		AuthMW: middleware.AuthConfig{JWT: jwtSvc, Repo: repos},
		Activity: middleware.ActivityUpdateConfig{Repo: repos},
		Cfg: cfg,
	})
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)
	return &testEnv{echo: e, repos: repos, auth: authUC, hub: hub, srv: srv}
}

func loginAccessToken(t *testing.T, env *testEnv, email, password string) string {
	t.Helper()
	out, err := env.auth.Login(t.Context(), usecase.LoginInput{Email: email, Password: password})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return out.AccessToken
}

func dialWS(t *testing.T, env *testEnv, accessToken string) *gorillaws.Conn {
	t.Helper()
	conn, err := dialWSWithOrigin(t, env, accessToken, "")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func dialWSWithOrigin(t *testing.T, env *testEnv, accessToken, origin string) (*gorillaws.Conn, error) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(env.srv.URL, "http") + "/ws"
	header := http.Header{}
	if accessToken != "" {
		header.Set("Cookie", middleware.AccessCookieName+"="+accessToken)
	}
	if origin != "" {
		header.Set("Origin", origin)
	}
	conn, _, err := gorillaws.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn, nil
}

func wsAuthConn(t *testing.T, env *testEnv, email, password string) *gorillaws.Conn {
	t.Helper()
	testutil.CreateUser(t, env.repos, strings.Split(email, "@")[0], email, password)
	conn := dialWS(t, env, loginAccessToken(t, env, email, password))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	return conn
}

func matchTwoUsers(t *testing.T, env *testEnv, aID, bID uuid.UUID) uuid.UUID {
	t.Helper()
	match := testutil.NewMatchingUC(env.repos)
	if _, err := match.Start(t.Context(), aID); err != nil {
		t.Fatalf("start a: %v", err)
	}
	if _, err := match.Start(t.Context(), bID); err != nil {
		t.Fatalf("start b: %v", err)
	}
	status, err := match.Status(t.Context(), aID)
	if err != nil || status.GameID == nil {
		t.Fatalf("status: %+v err=%v", status, err)
	}
	return *status.GameID
}

func setBothSecrets(t *testing.T, env *testEnv, gameID, aID, bID uuid.UUID, secretA, secretB string) {
	t.Helper()
	gameUC := testutil.NewGameUCWithNotifier(t, env.repos, env.hub)
	if err := gameUC.SetSecretNumber(t.Context(), aID, gameID, secretA); err != nil {
		t.Fatalf("secret a: %v", err)
	}
	if err := gameUC.SetSecretNumber(t.Context(), bID, gameID, secretB); err != nil {
		t.Fatalf("secret b: %v", err)
	}
}

func newMiniredisStore(t *testing.T) (*infrredis.Store, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return infrredis.NewStore(rdb), mr
}

func wsUserRedisKey(userID uuid.UUID) string {
	return fmt.Sprintf("ws:user:%s", userID)
}

func readWSMessage(t *testing.T, conn *gorillaws.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var msg map[string]any
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("json: %v raw=%s", err, raw)
	}
	return msg
}

func sendWS(t *testing.T, conn *gorillaws.Conn, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := conn.WriteMessage(gorillaws.TextMessage, b); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// §18.7 WebSocket
func TestWSRequiresAuthFirst(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	token := loginAccessToken(t, env, "alice@test.local", "password123")
	conn := dialWS(t, env, token)

	sendWS(t, conn, map[string]string{"type": "PING"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["message"] != "認証が必要です" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSAuthAndPingPong(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	token := loginAccessToken(t, env, "alice@test.local", "password123")
	conn := dialWS(t, env, token)

	sendWS(t, conn, map[string]string{"type": "AUTH"})
	msg := readWSMessage(t, conn)
	if msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}

	sendWS(t, conn, map[string]string{"type": "PING"})
	msg = readWSMessage(t, conn)
	if msg["type"] != "PONG" {
		t.Fatalf("pong: %+v", msg)
	}
}

func TestWSSetSecretViaUseCase(t *testing.T) {
	env := setupWSTest(t)
	match := testutil.NewMatchingUC(env.repos)

	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	if _, err := match.Start(t.Context(), a.ID); err != nil {
		t.Fatalf("start a: %v", err)
	}
	if _, err := match.Start(t.Context(), b.ID); err != nil {
		t.Fatalf("start b: %v", err)
	}
	status, err := match.Status(t.Context(), a.ID)
	if err != nil || status.GameID == nil {
		t.Fatalf("status: %+v err=%v", status, err)
	}
	gameUC := testutil.NewGameUCWithNotifier(t, env.repos, env.hub)
	if err := gameUC.SetSecretNumber(t.Context(), a.ID, *status.GameID, "1234"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	game, err := env.repos.Game.FindByID(t.Context(), *status.GameID)
	if err != nil || game == nil || game.Player1Secret == "" {
		t.Fatalf("secret saved: %+v err=%v", game, err)
	}
}
