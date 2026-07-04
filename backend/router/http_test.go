package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/router"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

type httpTestEnv struct {
	echo  *echo.Echo
	repos repository.Repos
}

func setupTestEcho(t *testing.T) *httpTestEnv {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	authUC := testutil.NewAuthUC(t, repos)
	gameUC := testutil.NewGameUC(t, repos)
	matchingUC := testutil.NewMatchingUC(repos)
	rankingUC := testutil.NewRankingUC(repos)
	adminUC := testutil.NewAdminUC(repos, rankingUC)

	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
	})
	router.Register(e, router.Deps{
		Auth: authUC, Profile: usecase.NewProfileUseCase(repos), Matching: matchingUC, Game: gameUC,
		Ranking: rankingUC, Admin: adminUC, JWT: jwtSvc,
		AuthMW: middleware.AuthConfig{JWT: jwtSvc, Repo: repos},
		Activity: middleware.ActivityUpdateConfig{Repo: repos},
		Cfg: cfg,
	})
	return &httpTestEnv{echo: e, repos: repos}
}

func (env *httpTestEnv) seedUser(t *testing.T, username, email, password string) {
	t.Helper()
	testutil.CreateUser(t, env.repos, username, email, password)
}

func (env *httpTestEnv) loginDirect(t *testing.T, email, password string) []*http.Cookie {
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

func (env *httpTestEnv) register(t *testing.T, username, email, password string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	env.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s status %d: %s", username, rec.Code, rec.Body.String())
	}
}

func (env *httpTestEnv) login(t *testing.T, email, password string) []*http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	env.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status %d: %s", rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()
}

func (env *httpTestEnv) do(t *testing.T, method, path string, cookies []*http.Cookie, body any) *httptest.ResponseRecorder {
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

// §18.6 API 結合テスト
func TestHealthOK(t *testing.T) {
	env := setupTestEcho(t)
	rec := env.do(t, http.MethodGet, "/health", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterAndLogin(t *testing.T) {
	env := setupTestEcho(t)
	env.register(t, "alice", "alice@test.local", "password123")

	dupBody, _ := json.Marshal(map[string]string{
		"username": "alice", "email": "alice@test.local", "password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(dupBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	env.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status %d", rec.Code)
	}

	cookies := env.login(t, "alice@test.local", "password123")
	if len(cookies) == 0 || cookies[0].Name != middleware.AccessCookieName {
		t.Fatalf("expected access_token cookie, got %v", cookies)
	}
}

func TestMeUnauthorized(t *testing.T) {
	env := setupTestEcho(t)
	rec := env.do(t, http.MethodGet, "/api/me", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestMeWithCookie(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	cookies := env.loginDirect(t, "alice@test.local", "password123")
	rec := env.do(t, http.MethodGet, "/api/me", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshAndLogout(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	cookies := env.loginDirect(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodPost, "/api/auth/refresh", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status %d: %s", rec.Code, rec.Body.String())
	}
	newCookies := rec.Result().Cookies()

	rec = env.do(t, http.MethodPost, "/api/auth/logout", newCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/me", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout status %d", rec.Code)
	}
}

func TestMatchingFlow(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	env.seedUser(t, "bob", "bob@test.local", "password123")
	aliceCookies := env.loginDirect(t, "alice@test.local", "password123")
	bobCookies := env.loginDirect(t, "bob@test.local", "password123")

	rec := env.do(t, http.MethodPost, "/api/matching/start", aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("alice start status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodPost, "/api/matching/start", bobCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("bob start status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/matching/status", aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Status string `json:"status"`
			GameID string `json:"gameId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Data.Status != "matched" || body.Data.GameID == "" {
		t.Fatalf("matched status: %+v", body.Data)
	}
}

func TestRankingEndpoint(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	cookies := env.loginDirect(t, "alice@test.local", "password123")
	rec := env.do(t, http.MethodGet, "/api/ranking", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("ranking status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminEndpoints(t *testing.T) {
	env := setupTestEcho(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	env.seedUser(t, "alice", "alice@test.local", "password123")
	adminCookies := env.loginDirect(t, "admin@test.local", "adminpass123")

	rec := env.do(t, http.MethodGet, "/api/admin/users", adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list users status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodPost, "/api/admin/ranking/rebuild", adminCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("rebuild ranking status %d: %s", rec.Code, rec.Body.String())
	}

	userCookies := env.loginDirect(t, "alice@test.local", "password123")
	rec = env.do(t, http.MethodGet, "/api/admin/users", userCookies, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin status %d", rec.Code)
	}
}

func TestGetGameEndpoint(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	env.seedUser(t, "bob", "bob@test.local", "password123")
	aliceCookies := env.loginDirect(t, "alice@test.local", "password123")
	bobCookies := env.loginDirect(t, "bob@test.local", "password123")

	env.do(t, http.MethodPost, "/api/matching/start", aliceCookies, nil)
	env.do(t, http.MethodPost, "/api/matching/start", bobCookies, nil)
	statusRec := env.do(t, http.MethodGet, "/api/matching/status", aliceCookies, nil)
	var statusBody struct {
		Data struct {
			GameID string `json:"gameId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("json: %v", err)
	}

	rec := env.do(t, http.MethodGet, "/api/games/"+statusBody.Data.GameID, aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("game status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/games/"+uuid.New().String(), aliceCookies, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing game status %d", rec.Code)
	}
}

func TestAdminDeleteUser(t *testing.T) {
	env := setupTestEcho(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.loginDirect(t, "admin@test.local", "adminpass123")
	user := testutil.CreateUser(t, env.repos, "victim", "victim@test.local", "password123")

	rec := env.do(t, http.MethodDelete, "/api/admin/users/"+user.ID.String(), adminCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete user status %d: %s", rec.Code, rec.Body.String())
	}
}
