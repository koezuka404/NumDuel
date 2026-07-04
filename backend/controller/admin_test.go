package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/numduel/numduel/testutil"
)

func TestAdminEndpoints(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	alice := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")

	rec := env.do(t, http.MethodGet, "/api/admin/users", adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list users status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/admin/users/search?q=ali", adminCookies, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("search users on sqlite status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodPost, "/api/admin/ranking/rebuild", adminCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("rebuild ranking status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/admin/logs/types", adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("log types status %d: %s", rec.Code, rec.Body.String())
	}

	from := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	to := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	rec = env.do(t, http.MethodGet, "/api/admin/logs?logType=login&userId="+alice.ID.String()+"&from="+from+"&to="+to, adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search logs status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/admin/logs/download?logType=login", adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("download logs status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodGet, "/api/admin/backup/status", adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("backup status %d: %s", rec.Code, rec.Body.String())
	}

	rec = env.do(t, http.MethodDelete, "/api/admin/users/"+alice.ID.String(), adminCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete user status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminForbiddenAndBadRequest(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	userCookies := env.login(t, "alice@test.local", "password123")

	rec := env.do(t, http.MethodGet, "/api/admin/users", userCookies, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin status %d", rec.Code)
	}

	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")

	rec = env.do(t, http.MethodGet, "/api/admin/logs?userId=bad-uuid", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad userId status %d", rec.Code)
	}
	rec = env.do(t, http.MethodGet, "/api/admin/logs?from=not-time", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad from status %d", rec.Code)
	}
	rec = env.do(t, http.MethodGet, "/api/admin/logs?to=not-time", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad to status %d", rec.Code)
	}
	rec = env.do(t, http.MethodDelete, "/api/admin/users/not-uuid", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad delete id status %d", rec.Code)
	}
}

func TestAdminSearchUsersEmptyQuery(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")
	rec := env.do(t, http.MethodGet, "/api/admin/users/search", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty search status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodGet, "/api/admin/users", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("admin unauthorized status %d", rec.Code)
	}
}

func TestAdminRebuildRankingUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/admin/ranking/rebuild", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("rebuild unauthorized status %d", rec.Code)
	}
}

func TestAdminDownloadLogsUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodGet, "/api/admin/logs/download", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("download unauthorized status %d", rec.Code)
	}
}

func TestAdminDeleteSelfForbidden(t *testing.T) {
	env := setupCtrlEnv(t)
	master := testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")
	rec := env.do(t, http.MethodDelete, "/api/admin/users/"+master.ID.String(), adminCookies, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("delete self status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminDeleteUserNotFound(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")
	rec := env.do(t, http.MethodDelete, "/api/admin/users/"+uuid.New().String(), adminCookies, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing user status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSearchLogsDownloadBadTo(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")
	rec := env.do(t, http.MethodGet, "/api/admin/logs/download?to=bad", adminCookies, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("download bad to status %d", rec.Code)
	}
}

func TestAdminDownloadBadQuery(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.SeedMaster(t, env.repos, "admin@test.local", "adminpass123")
	adminCookies := env.login(t, "admin@test.local", "adminpass123")
	rec := env.do(t, http.MethodGet, "/api/admin/logs/download?userId="+uuid.New().String()+"&from="+time.Now().UTC().Format(time.RFC3339), adminCookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("download with filters status %d", rec.Code)
	}
}
