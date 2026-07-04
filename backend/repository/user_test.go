package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestUserRepoCRUD(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()

	user := createUser(t, repos, "alice", "alice@test.local")
	user.WinCount = 3
	if err := repos.User.Update(ctx, user); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repos.User.FindByID(ctx, user.ID)
	if err != nil || got == nil || got.WinCount != 3 {
		t.Fatalf("find by id: %+v err=%v", got, err)
	}

	got, err = repos.User.FindByEmail(ctx, "alice@test.local")
	if err != nil || got == nil || got.ID != user.ID {
		t.Fatalf("find by email: %+v err=%v", got, err)
	}

	got, err = repos.User.FindByUsername(ctx, "alice")
	if err != nil || got == nil || got.ID != user.ID {
		t.Fatalf("find by username: %+v err=%v", got, err)
	}

	missing, err := repos.User.FindByID(ctx, uuid.New())
	if err != nil || missing != nil {
		t.Fatalf("missing id: %+v err=%v", missing, err)
	}
}

func TestUserRepoListAndSearch(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	createUser(t, repos, "alice", "alice@test.local")
	createUser(t, repos, "bob", "bob@test.local")

	all, err := repos.User.ListAll(ctx)
	if err != nil || len(all) < 2 {
		t.Fatalf("list all: %d err=%v", len(all), err)
	}

	page, total, err := repos.User.List(ctx, 0, 0)
	if err != nil || total < 2 || len(page) == 0 {
		t.Fatalf("list paginated: len=%d total=%d err=%v", len(page), total, err)
	}

	found, total, err := repos.User.Search(ctx, "ali", 1, 10)
	if err != nil || total == 0 || len(found) == 0 {
		t.Fatalf("search: len=%d total=%d err=%v", len(found), total, err)
	}

	page2, total2, err := repos.User.List(ctx, 2, 1)
	if err != nil || total2 < 2 {
		t.Fatalf("list page 2: len=%d total=%d err=%v", len(page2), total2, err)
	}
}

func TestUserRepoActivityAndMaster(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()

	exists, err := repos.User.ExistsActiveMaster(ctx)
	if err != nil || exists {
		t.Fatalf("no master yet: exists=%v err=%v", exists, err)
	}

	user := createUser(t, repos, "alice", "alice@test.local")
	oldActivity := user.LastActivityAt
	touchAt := time.Now().UTC().Add(time.Hour)
	if err := repos.User.TouchLastActivity(ctx, user.ID, touchAt); err != nil {
		t.Fatalf("touch activity: %v", err)
	}
	got, _ := repos.User.FindByID(ctx, user.ID)
	if !got.LastActivityAt.After(oldActivity) {
		t.Fatalf("activity not updated: %v", got.LastActivityAt)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.User.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) == 0 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}

	before := time.Now().UTC().Add(time.Hour)
	inactive, err := repos.User.ListInactiveSince(ctx, before)
	if err != nil || len(inactive) == 0 {
		t.Fatalf("inactive since: %d err=%v", len(inactive), err)
	}

	master := createUser(t, repos, "admin", "admin@test.local")
	master.Role = model.RoleMaster
	if err := repos.User.Update(ctx, master); err != nil {
		t.Fatalf("promote master: %v", err)
	}
	exists, err = repos.User.ExistsActiveMaster(ctx)
	if err != nil || !exists {
		t.Fatalf("master exists: %v err=%v", exists, err)
	}
}
