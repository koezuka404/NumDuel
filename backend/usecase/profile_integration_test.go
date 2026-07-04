package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestGetProfile(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	now := time.Now().UTC()

	if err := repos.Ranking.ReplaceAll(context.Background(), []model.Ranking{
		{UserID: user.ID, Rank: 1, Username: "alice", WinCount: 2, UpdatedAt: now},
	}); err != nil {
		t.Fatalf("ranking: %v", err)
	}

	out, err := profile.GetProfile(context.Background(), user.ID)
	if err != nil || out.Username != "alice" || out.WinCount != 0 || out.Rank == nil || *out.Rank != 1 {
		t.Fatalf("profile: %+v err=%v", out, err)
	}
}

func TestGetProfileUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)

	_, err := profile.GetProfile(context.Background(), uuid.New())
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("missing user: %v", err)
	}

	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	now := time.Now().UTC()
	user.DeletedAt = &now
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	_, err = profile.GetProfile(context.Background(), user.ID)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("deleted user: %v", err)
	}
}

func TestGetMatchHistory(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	opponent := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	now := time.Now().UTC()
	gameID := uuid.New()

	if err := repos.MatchHistory.Create(context.Background(), &model.MatchHistory{
		ID: uuid.New(), GameID: gameID, WinnerID: user.ID, LoserID: opponent.ID,
		WinnerUsername: "alice", LoserUsername: "bob", FinishedAt: now, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("match history: %v", err)
	}

	items, total, err := profile.GetMatchHistory(context.Background(), user.ID, 1, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].GameID != gameID {
		t.Fatalf("history: items=%+v total=%d err=%v", items, total, err)
	}
}

func TestGetLoginHistory(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	now := time.Now().UTC()

	if err := repos.LoginLog.Create(context.Background(), &model.LoginLog{
		ID: uuid.New(), UserID: user.ID, Action: model.LoginActionLogin, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("login log: %v", err)
	}

	items, total, err := profile.GetLoginHistory(context.Background(), user.ID, 1, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].Action != model.LoginActionLogin {
		t.Fatalf("login history: items=%+v total=%d err=%v", items, total, err)
	}
}

func TestGetWSHistory(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	now := time.Now().UTC()
	disconnected := now.Add(time.Minute)

	if err := repos.WSConnectionLog.Create(context.Background(), &model.WSConnectionLog{
		ID: uuid.New(), UserID: user.ID, ConnectionID: "conn-1", ConnectedAt: now, DisconnectedAt: &disconnected,
	}); err != nil {
		t.Fatalf("ws log: %v", err)
	}

	items, total, err := profile.GetWSHistory(context.Background(), user.ID, 1, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ConnectionID != "conn-1" {
		t.Fatalf("ws history: items=%+v total=%d err=%v", items, total, err)
	}
}

func TestGetMatchHistoryUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)

	_, _, err := profile.GetMatchHistory(context.Background(), uuid.New(), 1, 10)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("unauthorized: %v", err)
	}
}

func TestGetLoginHistoryUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)

	_, _, err := profile.GetLoginHistory(context.Background(), uuid.New(), 1, 10)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("unauthorized: %v", err)
	}
}

func TestGetWSHistoryUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)

	_, _, err := profile.GetWSHistory(context.Background(), uuid.New(), 1, 10)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("unauthorized: %v", err)
	}
}
