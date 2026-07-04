package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/numduel/numduel/testutil"
)

func TestGameGet(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	aliceCookies := env.login(t, "alice@test.local", "password123")
	bobCookies := env.login(t, "bob@test.local", "password123")

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
		t.Fatalf("get game status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGameGetForbidden(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	testutil.CreateUser(t, env.repos, "carol", "carol@test.local", "password123")
	aliceCookies := env.login(t, "alice@test.local", "password123")
	bobCookies := env.login(t, "bob@test.local", "password123")
	carolCookies := env.login(t, "carol@test.local", "password123")

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
		t.Fatalf("forbidden status %d", rec.Code)
	}
}

func TestGameBadRequestAndUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	cookies := env.login(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodGet, "/api/games/not-a-uuid", cookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id status %d", rec.Code)
	}

	rec = env.do(t, http.MethodGet, "/api/games/"+uuid.New().String(), cookies, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing game status %d", rec.Code)
	}

	rec = env.do(t, http.MethodGet, "/api/games/"+uuid.New().String(), nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status %d", rec.Code)
	}
}
