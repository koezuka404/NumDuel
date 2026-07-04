package controller_test

import (
	"net/http"
	"testing"

	"github.com/numduel/numduel/testutil"
)

func TestMeEndpoints(t *testing.T) {
	env := setupCtrlEnv(t)
	user := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	env.seedHistories(t, user.ID)
	cookies := env.login(t, "alice@test.local", "password123")

	endpoints := []string{
		"/api/me",
		"/api/me/profile",
		"/api/me/match-history",
		"/api/me/login-history",
		"/api/me/ws-history",
	}
	for _, path := range endpoints {
		rec := env.do(t, http.MethodGet, path, cookies, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status %d: %s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestMeUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	paths := []string{"/api/me", "/api/me/profile", "/api/me/match-history", "/api/me/login-history", "/api/me/ws-history"}
	for _, path := range paths {
		rec := env.do(t, http.MethodGet, path, nil, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s unauthorized status %d", path, rec.Code)
		}
	}
}

func TestMePagedQuery(t *testing.T) {
	env := setupCtrlEnv(t)
	user := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	env.seedHistories(t, user.ID)
	cookies := env.login(t, "alice@test.local", "password123")
	rec := env.do(t, http.MethodGet, "/api/me/match-history?page=1&limit=5", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("paged history status %d: %s", rec.Code, rec.Body.String())
	}
}
