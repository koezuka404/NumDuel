package controller_test

import (
	"net/http"
	"testing"

	"github.com/numduel/numduel/testutil"
)

func TestRankingGet(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	cookies := env.login(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodGet, "/api/ranking", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("ranking status %d: %s", rec.Code, rec.Body.String())
	}
}
