package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// §18.5.2 マッチング系
func TestStartMatchingWaiting(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	out, err := match.Start(context.Background(), user.ID)
	if err != nil || out.Status != "waiting" {
		t.Fatalf("start: %+v err=%v", out, err)
	}

	_, err = match.Start(context.Background(), user.ID)
	if !errors.Is(err, usecase.ErrAlreadyInMatching) {
		t.Fatalf("already in matching: %v", err)
	}
}

func TestMatchTwoPlayersStartReturnsMatched(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")

	outA, err := match.Start(context.Background(), a.ID)
	if err != nil || outA.Status != "waiting" || outA.GameID != nil {
		t.Fatalf("alice start: %+v err=%v", outA, err)
	}
	outB, err := match.Start(context.Background(), b.ID)
	if err != nil || outB.Status != "matched" || outB.GameID == nil {
		t.Fatalf("bob start: %+v err=%v", outB, err)
	}
}

func TestMatchTwoPlayers(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")

	if _, err := match.Start(context.Background(), a.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	if _, err := match.Start(context.Background(), b.ID); err != nil {
		t.Fatalf("bob start: %v", err)
	}

	statusA, err := match.Status(context.Background(), a.ID)
	if err != nil || statusA.Status != "matched" || statusA.GameID == nil {
		t.Fatalf("alice status: %+v err=%v", statusA, err)
	}
	statusB, err := match.Status(context.Background(), b.ID)
	if err != nil || statusB.GameID == nil || *statusB.GameID != *statusA.GameID {
		t.Fatalf("bob status: %+v err=%v", statusB, err)
	}

	game, err := repos.Game.FindByID(context.Background(), *statusA.GameID)
	if err != nil || game == nil {
		t.Fatalf("game: %v", err)
	}
	if game.Player1ID != a.ID {
		t.Fatalf("player1 = first waiter, got %v want %v", game.Player1ID, a.ID)
	}
}

func TestCancelMatching(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if _, err := match.Start(context.Background(), user.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
	out, err := match.Cancel(context.Background(), user.ID)
	if err != nil || out.Status != "cancelled" {
		t.Fatalf("cancel: %+v err=%v", out, err)
	}
	status, err := match.Status(context.Background(), user.ID)
	if err != nil || status.Status != "idle" {
		t.Fatalf("status after cancel: %+v err=%v", status, err)
	}
}

func TestMasterCannotMatch(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	master := seedMaster(t, repos)

	_, err := match.Start(context.Background(), master.ID)
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("master forbidden: %v", err)
	}
}

func seedMaster(t *testing.T, repos repository.Repos) *model.User {
	t.Helper()
	now := time.Now().UTC()
	user := &model.User{
		ID: uuid.New(), Username: "admin", Email: "admin@test.local",
		PasswordHash: "hash", Role: model.RoleMaster,
		LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.User.Create(context.Background(), user); err != nil {
		t.Fatalf("create master: %v", err)
	}
	return user
}
