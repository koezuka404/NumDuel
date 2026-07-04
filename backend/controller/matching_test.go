package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/numduel/numduel/testutil"
)

func TestMatchingEndpoints(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	aliceCookies := env.login(t, "alice@test.local", "password123")
	bobCookies := env.login(t, "bob@test.local", "password123")

	rec := env.do(t, http.MethodPost, "/api/matching/start", aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("start alice status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/matching/status", aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status waiting %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodPost, "/api/matching/start", bobCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("start bob status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/matching/status", aliceCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status matched %d: %s", rec.Code, rec.Body.String())
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
		t.Fatalf("matched body: %+v", body.Data)
	}
}

func TestMatchingCancel(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	cookies := env.login(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodPost, "/api/matching/start", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("start status %d", rec.Code)
	}
	rec = env.do(t, http.MethodPost, "/api/matching/cancel", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMatchingCancelNotInQueue(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	cookies := env.login(t, "alice@test.local", "password123")
	rec := env.do(t, http.MethodPost, "/api/matching/cancel", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel not in queue status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMatchingStartWhileAlreadyMatched(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	aliceCookies := env.login(t, "alice@test.local", "password123")
	bobCookies := env.login(t, "bob@test.local", "password123")
	env.do(t, http.MethodPost, "/api/matching/start", aliceCookies, nil)
	env.do(t, http.MethodPost, "/api/matching/start", bobCookies, nil)
	rec := env.do(t, http.MethodPost, "/api/matching/start", aliceCookies, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("start while matched status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMatchingUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	for _, path := range []string{"/api/matching/start", "/api/matching/cancel", "/api/matching/status"} {
		method := http.MethodPost
		if path == "/api/matching/status" {
			method = http.MethodGet
		}
		rec := env.do(t, method, path, nil, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s unauthorized status %d", path, rec.Code)
		}
	}
}
