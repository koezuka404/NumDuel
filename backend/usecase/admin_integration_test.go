package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// §18.5.4 管理系
func TestDeleteUserCannotDeleteSelf(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	err := admin.DeleteUser(context.Background(), master.ID, master.ID)
	if !errors.Is(err, usecase.ErrCannotDeleteSelf) {
		t.Fatalf("delete self: %v", err)
	}
}

func TestDeleteUserCannotDeleteMaster(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	otherMaster := testutil.CreateUser(t, repos, "master2", "master2@test.local", "password123")
	otherMaster.Role = model.RoleMaster
	if err := repos.User.Update(context.Background(), otherMaster); err != nil {
		t.Fatalf("promote master: %v", err)
	}

	err := admin.DeleteUser(context.Background(), master.ID, otherMaster.ID)
	if !errors.Is(err, usecase.ErrCannotDeleteMaster) {
		t.Fatalf("delete master: %v", err)
	}
}

func TestDeleteUserInActiveGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	match := testutil.NewMatchingUC(repos)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	matchTwo(t, match, a.ID, b.ID)

	err := admin.DeleteUser(context.Background(), master.ID, a.ID)
	if !errors.Is(err, usecase.ErrUserInActiveGame) {
		t.Fatalf("active game: %v", err)
	}
}

func TestDeleteUserSuccess(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	target := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if err := admin.DeleteUser(context.Background(), master.ID, target.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	user, err := repos.User.FindByID(context.Background(), target.ID)
	if err != nil || user.DeletedAt == nil {
		t.Fatalf("deleted user: %+v err=%v", user, err)
	}
}

func TestRebuildRankingExcludesMaster(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	a.WinCount = 3
	b.WinCount = 1
	master.WinCount = 99
	ctx := context.Background()
	if err := repos.User.Update(ctx, a); err != nil {
		t.Fatalf("update alice: %v", err)
	}
	if err := repos.User.Update(ctx, b); err != nil {
		t.Fatalf("update bob: %v", err)
	}
	if err := repos.User.Update(ctx, master); err != nil {
		t.Fatalf("update master: %v", err)
	}

	if err := admin.RebuildRanking(ctx, master.ID); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	rows, err := ranking.Get(ctx)
	if err != nil {
		t.Fatalf("get ranking: %v", err)
	}
	if len(rows) != 2 || rows[0].Username != "alice" || rows[1].Username != "bob" {
		t.Fatalf("ranking: %+v", rows)
	}
	for _, row := range rows {
		if row.Username == "admin" {
			t.Fatalf("master should be excluded")
		}
	}
}

func TestAdminSearchUsersRequiresQuery(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	_, err := admin.SearchUsers(context.Background(), "")
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("empty query: %v", err)
	}
}

func TestAdminDownloadActivityLogsCSV(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	now := time.Now().UTC()
	detail, _ := json.Marshal(map[string]string{"action": "test"})
	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), LogType: "admin_test", Detail: detail, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create log: %v", err)
	}
	csv, err := admin.DownloadActivityLogsCSV(context.Background(), master.ID, "", nil, nil, nil)
	if err != nil || len(csv) == 0 {
		t.Fatalf("csv: len=%d err=%v", len(csv), err)
	}
}
