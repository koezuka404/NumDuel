package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// セキュリティ: 失効・強制ログアウト後のトークン拒否
func TestWSAuthRejectsRevokedToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	revoker := testutil.NewMemJWTRevoker()
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := jwtSvc.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := revoker.Revoke(context.Background(), claims.JTI, time.Hour); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, revoker, nil, nil, nil)
	_, err = wsAuth.Authenticate(context.Background(), token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("revoked token: %v", err)
	}
}

func TestWSAuthRejectsForceLogoutToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	force := testutil.NewMemForceLogout()
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	issuedAt := time.Now().UTC().Add(-30 * time.Minute)
	token, err := jwtSvc.Issue(user.ID, user.Role, issuedAt)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if err := force.SetForceLogoutBefore(context.Background(), user.ID, time.Now().UTC()); err != nil {
		t.Fatalf("force logout: %v", err)
	}

	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, force, nil, nil)
	_, err = wsAuth.Authenticate(context.Background(), token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("force logout token: %v", err)
	}
}

func TestWSAuthRejectsEmptyToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil, nil)
	_, err = wsAuth.Authenticate(context.Background(), "")
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("empty token: %v", err)
	}
}

func TestWSAuthRejectsDeletedUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	now := time.Now().UTC()
	user.DeletedAt = &now
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil, nil)
	_, err = wsAuth.Authenticate(context.Background(), token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("deleted user token: %v", err)
	}
}

func TestLogoutRevokesAccessTokenJTI(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	revoker := testutil.NewMemJWTRevoker()
	auth := testutil.NewAuthUCWithRevoker(t, repos, revoker)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	claims, err := jwtSvc.Parse(login.AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := auth.Logout(context.Background(), usecase.LogoutInput{
		UserID: user.ID, JTI: claims.JTI, Exp: claims.ExpiresAt,
	}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	revoked, err := revoker.IsRevoked(context.Background(), claims.JTI)
	if err != nil || !revoked {
		t.Fatalf("jti not revoked: revoked=%v err=%v", revoked, err)
	}
}
