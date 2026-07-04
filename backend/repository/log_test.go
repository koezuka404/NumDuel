package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestActivityLogRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	user := createUser(t, repos, "alice", "alice@test.local")

	log1 := newActivityLog("login", &user.ID)
	log2 := newActivityLog("game", nil)
	for _, l := range []*model.ActivityLog{log1, log2} {
		if err := repos.ActivityLog.Create(ctx, l); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC().Add(time.Hour)
	rows, total, err := repos.ActivityLog.Search(ctx, "login", &user.ID, &from, &to, 1, 10)
	if err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("search filtered: total=%d len=%d err=%v", total, len(rows), err)
	}

	rows, total, err = repos.ActivityLog.Search(ctx, "", nil, nil, nil, 0, 0)
	if err != nil || total < 2 {
		t.Fatalf("search all: total=%d err=%v", total, err)
	}

	types, err := repos.ActivityLog.ListDistinctLogTypes(ctx)
	if err != nil || len(types) < 2 {
		t.Fatalf("log types: %v err=%v", types, err)
	}

	old := newActivityLog("old", nil)
	old.CreatedAt = time.Now().UTC().Add(-48 * time.Hour)
	old.UpdatedAt = old.CreatedAt
	if err := repos.ActivityLog.Create(ctx, old); err != nil {
		t.Fatalf("create old: %v", err)
	}
	n, err := repos.ActivityLog.DeleteOlderThan(ctx, time.Now().UTC().Add(-24*time.Hour), 10)
	if err != nil || n == 0 {
		t.Fatalf("delete older: n=%d err=%v", n, err)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.ActivityLog.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) == 0 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}

func TestActivityLogRepoSearchFilters(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	user := createUser(t, repos, "alice", "alice@test.local")

	log := newActivityLog("game", &user.ID)
	if err := repos.ActivityLog.Create(ctx, log); err != nil {
		t.Fatalf("create: %v", err)
	}

	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC().Add(time.Hour)

	cases := []struct {
		name    string
		logType string
		userID  *uuid.UUID
		from    *time.Time
		to      *time.Time
	}{
		{"logType only", "game", nil, nil, nil},
		{"user only", "", &user.ID, nil, nil},
		{"from only", "", nil, &from, nil},
		{"to only", "", nil, nil, &to},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, total, err := repos.ActivityLog.Search(ctx, tc.logType, tc.userID, tc.from, tc.to, 1, 10)
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			if total == 0 || len(rows) == 0 {
				t.Fatalf("expected results: total=%d len=%d", total, len(rows))
			}
		})
	}
}

func TestLoginLogRepoListPaged(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	user := createUser(t, repos, "alice", "alice@test.local")

	log := newLoginLog(user.ID, model.LoginActionLogin)
	if err := repos.LoginLog.Create(ctx, log); err != nil {
		t.Fatalf("create: %v", err)
	}

	rows, total, err := repos.LoginLog.ListByUserID(ctx, user.ID, 0, 0)
	if err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("list: total=%d len=%d err=%v", total, len(rows), err)
	}

	old := newLoginLog(user.ID, model.LoginActionLogout)
	old.CreatedAt = time.Now().UTC().Add(-48 * time.Hour)
	old.UpdatedAt = old.CreatedAt
	if err := repos.LoginLog.Create(ctx, old); err != nil {
		t.Fatalf("create old: %v", err)
	}
	n, err := repos.LoginLog.DeleteOlderThan(ctx, time.Now().UTC().Add(-24*time.Hour), 5)
	if err != nil || n == 0 {
		t.Fatalf("delete older: n=%d err=%v", n, err)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.LoginLog.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) == 0 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}

func TestWSConnectionLogRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	user := createUser(t, repos, "alice", "alice@test.local")

	log := newWSLog(user.ID)
	if err := repos.WSConnectionLog.Create(ctx, log); err != nil {
		t.Fatalf("create: %v", err)
	}

	disconnected := time.Now().UTC()
	if err := repos.WSConnectionLog.UpdateDisconnected(ctx, log.ID, disconnected); err != nil {
		t.Fatalf("update disconnected: %v", err)
	}

	rows, total, err := repos.WSConnectionLog.ListByUserID(ctx, user.ID, 1, 10)
	if err != nil || total != 1 || len(rows) != 1 || rows[0].DisconnectedAt == nil {
		t.Fatalf("list: %+v total=%d err=%v", rows, total, err)
	}

	old := newWSLog(user.ID)
	old.ConnectedAt = time.Now().UTC().Add(-48 * time.Hour)
	if err := repos.WSConnectionLog.Create(ctx, old); err != nil {
		t.Fatalf("create old: %v", err)
	}
	n, err := repos.WSConnectionLog.DeleteOlderThan(ctx, time.Now().UTC().Add(-24*time.Hour), 5)
	if err != nil || n == 0 {
		t.Fatalf("delete older: n=%d err=%v", n, err)
	}
}
