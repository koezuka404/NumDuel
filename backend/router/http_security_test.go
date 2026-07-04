package router_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/config"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/router"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func setupSecureEcho(t *testing.T) *httpTestEnv {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	revoker := testutil.NewMemJWTRevoker()
	authUC := testutil.NewAuthUCWithRevoker(t, repos, revoker)
	gameUC := testutil.NewGameUC(t, repos)
	matchingUC := testutil.NewMatchingUC(repos)
	rankingUC := testutil.NewRankingUC(repos)
	adminUC := testutil.NewAdminUC(repos, rankingUC)

	cfg := &config.Config{CookieSecure: false, JWTExpiryMinutes: 60, RefreshTokenExpiryDays: 7}
	e := echo.New()
	router.Register(e, router.Deps{
		Auth: authUC, Profile: usecase.NewProfileUseCase(repos), Matching: matchingUC, Game: gameUC,
		Ranking: rankingUC, Admin: adminUC, JWT: jwtSvc,
		AuthMW: middleware.AuthConfig{JWT: jwtSvc, Revoker: revoker, Repo: repos},
		Activity: middleware.ActivityUpdateConfig{Repo: repos},
		Cfg: cfg,
	})
	return &httpTestEnv{echo: e, repos: repos}
}

// セキュリティ: API 層の認可・情報漏洩防止
func TestGetGameForbiddenForNonParticipant(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	env.seedUser(t, "bob", "bob@test.local", "password123")
	env.seedUser(t, "carol", "carol@test.local", "password123")
	aliceCookies := env.loginDirect(t, "alice@test.local", "password123")
	bobCookies := env.loginDirect(t, "bob@test.local", "password123")
	carolCookies := env.loginDirect(t, "carol@test.local", "password123")

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

	rec := env.do(t, http.MethodGet, "/api/games/"+statusBody.Data.GameID, carolCookies, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-participant status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetGameResponseDoesNotLeakSecrets(t *testing.T) {
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
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, forbidden := range []string{"Player1Secret", "Player2Secret", "secretNumber", "password"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaks %q: %s", forbidden, body)
		}
	}
}

func TestMeRejectsRevokedTokenAfterLogout(t *testing.T) {
	env := setupSecureEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	cookies := env.loginDirect(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodPost, "/api/auth/logout", cookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status %d", rec.Code)
	}

	rec = env.do(t, http.MethodGet, "/api/me", cookies, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("me with revoked token status %d", rec.Code)
	}
}

func TestAdminDeleteOtherUserForbiddenForRegularUser(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	aliceCookies := env.loginDirect(t, "alice@test.local", "password123")
	victim := testutil.CreateUser(t, env.repos, "victim", "victim@test.local", "password123")

	rec := env.do(t, http.MethodDelete, "/api/admin/users/"+victim.ID.String(), aliceCookies, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("user delete admin status %d", rec.Code)
	}
}

func TestGetGameInvalidIDRejected(t *testing.T) {
	env := setupTestEcho(t)
	env.seedUser(t, "alice", "alice@test.local", "password123")
	cookies := env.loginDirect(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodGet, "/api/games/not-a-uuid", cookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid id status %d", rec.Code)
	}

	rec = env.do(t, http.MethodGet, "/api/games/"+uuid.New().String(), cookies, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing game status %d", rec.Code)
	}
}
