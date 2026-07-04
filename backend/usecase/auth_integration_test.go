package usecase_test

import (
	"context"
	"errors"
	"testing"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// §18.5.1 認証系
func TestRegisterUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	out, err := auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "alice@test.local", Password: "password123",
	})
	if err != nil || out.Username != "alice" {
		t.Fatalf("register: out=%+v err=%v", out, err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice2", Email: "alice@test.local", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrDuplicateUser) {
		t.Fatalf("duplicate email: %v", err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "ab", Email: "bad@test.local", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("validation: %v", err)
	}
}

func TestLoginUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	out, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil || out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("login: %+v err=%v", out, err)
	}

	_, err = auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "wrongpass1",
	})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("bad password: %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	refreshed, err := auth.Refresh(context.Background(), usecase.RefreshInput{
		RefreshToken: login.RefreshToken,
	})
	if err != nil || refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatalf("refresh: %+v err=%v", refreshed, err)
	}

	_, err = auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("reused refresh token: %v", err)
	}
}

func TestLogoutUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	claims, err := jwtSvc.Parse(login.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	if err := auth.Logout(context.Background(), usecase.LogoutInput{
		UserID: user.ID, JTI: claims.JTI, Exp: claims.ExpiresAt,
	}); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("refresh after logout: %v", err)
	}
}
