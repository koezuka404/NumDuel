package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestRefreshEmptyToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	_, err := auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "  "})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("empty refresh: %v", err)
	}
}

func TestRefreshExpiredToken(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	hash := infrcrypto.NewRefreshTokenService().Hash(login.RefreshToken)
	past := time.Now().UTC().Add(-48 * time.Hour)
	if err := gdb.Model(&model.RefreshToken{}).Where("token_hash = ?", hash).
		Update("expires_at", past).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}

	_, err = auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("expired refresh: %v", err)
	}

	stored, err := repos.RefreshToken.FindByTokenHash(context.Background(), hash)
	if err != nil || stored == nil || !stored.IsRevoked() {
		t.Fatalf("expired token should be revoked: %+v err=%v", stored, err)
	}
}

func TestRefreshDeletedUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	now := time.Now().UTC()
	user.DeletedAt = &now
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err = auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("deleted user refresh: %v", err)
	}
}

func TestRefreshReusedTokenRevokesFamilyAndWSSession(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	wsSessions := &memWSSessionStore{}
	auth := testutil.NewAuthUC(t, repos)
	auth.WSSessions = wsSessions
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	login, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	refreshed, err := auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	_ = refreshed

	_, err = auth.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("reused token: %v", err)
	}
	if !wsSessions.wasDeleted(user.ID) {
		t.Fatalf("ws session should be deleted on token reuse")
	}
}

func TestCleanupExpiredRefreshTokensDeletesOld(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	old := time.Now().UTC().AddDate(0, 0, -30)
	token := model.NewRefreshToken(user.ID, "old-hash", user.ID, old.Add(-time.Hour), old)
	token.Status = model.RefreshTokenRevoked
	token.RevokedAt = &old
	if err := repos.RefreshToken.Create(context.Background(), &token); err != nil {
		t.Fatalf("create old token: %v", err)
	}

	auth.CleanupExpiredRefreshTokens(context.Background())

	var count int64
	if err := gdb.Model(&model.RefreshToken{}).Where("token_hash = ?", "old-hash").Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("old token should be deleted, count=%d", count)
	}
}

func TestLogoutExpiredJWT(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	err := auth.Logout(context.Background(), usecase.LogoutInput{
		UserID: user.ID, JTI: "jti", Exp: time.Now().UTC().Add(-time.Minute),
	})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("expired jwt logout: %v", err)
	}
}

func TestLogoutEmptyJTI(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	err := auth.Logout(context.Background(), usecase.LogoutInput{
		UserID: user.ID, JTI: "", Exp: time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("empty jti logout: %v", err)
	}
}

func TestLoginDeletedUserByEmail(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	now := time.Now().UTC()
	user.DeletedAt = &now
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err := auth.Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("deleted user login: %v", err)
	}
}

func TestRegisterDuplicateUsernameOnly(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)

	_, err := auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = auth.Register(context.Background(), usecase.RegisterInput{
		Username: "alice", Email: "other@test.local", Password: "password123",
	})
	if !errors.Is(err, usecase.ErrDuplicateUser) {
		t.Fatalf("duplicate username: %v", err)
	}
}

func TestAuthNowAndCleanupGraceDefaults(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	fixed := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	auth.Now = func() time.Time { return fixed }
	auth.CleanupGrace = 0

	auth.CleanupExpiredRefreshTokens(context.Background())
}

func TestCleanupExpiredRefreshTokensDeletesMultiple(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	old := time.Now().UTC().AddDate(0, 0, -30)

	for i := 0; i < 3; i++ {
		token := model.NewRefreshToken(user.ID, fmt.Sprintf("hash-%d", i), user.ID, old.Add(-time.Hour), old)
		token.Status = model.RefreshTokenRevoked
		token.RevokedAt = &old
		if err := repos.RefreshToken.Create(context.Background(), &token); err != nil {
			t.Fatalf("create token: %v", err)
		}
	}

	auth.CleanupExpiredRefreshTokens(context.Background())

	var count int64
	if err := gdb.Model(&model.RefreshToken{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("old tokens should be deleted, count=%d", count)
	}
}

func TestLogoutRevokerError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auth := testutil.NewAuthUCWithRevoker(t, repos, failingJWTRevoker{})
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	err := auth.Logout(context.Background(), usecase.LogoutInput{
		UserID: user.ID, JTI: "jti-1", Exp: time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("revoker error: %v", err)
	}
}
