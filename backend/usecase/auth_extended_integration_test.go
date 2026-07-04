package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestGetMe(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	me, err := auth.GetMe(context.Background(), user.ID)
	if err != nil || me.Username != "alice" || me.Role != "user" {
		t.Fatalf("get me: %+v err=%v", me, err)
	}
}

func TestGetMeUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	_, err := auth.GetMe(context.Background(), uuid.New())
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("missing user: %v", err)
	}
}

func TestSeedMasterEmptyInput(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	if err := auth.SeedMaster(context.Background(), usecase.SeedMasterInput{}); err != nil {
		t.Fatalf("empty seed: %v", err)
	}
	master, err := repos.User.FindByUsername(context.Background(), "admin")
	if err != nil || master != nil {
		t.Fatalf("master should not exist: %+v err=%v", master, err)
	}
}

func TestSeedMasterSkipsWhenExists(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	if err := auth.SeedMaster(context.Background(), usecase.SeedMasterInput{
		Email: "other@test.local", Password: "otherpass123",
	}); err != nil {
		t.Fatalf("second seed: %v", err)
	}
}

func TestSeedMasterValidationErrors(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	err := auth.SeedMaster(context.Background(), usecase.SeedMasterInput{
		Email: "bad-email", Password: "adminpass123",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("bad email: %v", err)
	}

	err = auth.SeedMaster(context.Background(), usecase.SeedMasterInput{
		Email: "admin@test.local", Password: "short",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("weak password: %v", err)
	}
}

func TestRegisterValidationEdgeCases(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	_, err := auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("empty email: %v", err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "alice@test.local", Password: "short",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("weak password: %v", err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("valid register: %v", err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "other@test.local", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrDuplicateUser) {
		t.Fatalf("duplicate username: %v", err)
	}
}

func TestLoginValidationErrors(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	_, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "not-an-email", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("bad login email: %v", err)
	}

	_, err = auth.Login(context.Background(), usecase.LoginInput{
		Email: "user@test.local", Password: "short",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("weak login password: %v", err)
	}
}

func TestCleanupExpiredRefreshTokens(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	_ = login

	auth.CleanupExpiredRefreshTokens(context.Background())
}
