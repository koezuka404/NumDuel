package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestWSAuthSuccess(t *testing.T) {
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

	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil)
	out, err := wsAuth.Authenticate(context.Background(), token)
	if err != nil || out.UserID != user.ID {
		t.Fatalf("auth success: %+v err=%v", out, err)
	}
}

func TestWSAuthInvalidToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil)

	_, err := wsAuth.Authenticate(context.Background(), "not-a-jwt")
	if err == nil {
		t.Fatalf("expected invalid token error")
	}
}

func TestWSAuthCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil)
	fixed := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	wsAuth.Now = func() time.Time { return fixed }

	logID, err := wsAuth.RecordConnection(context.Background(), user.ID, "conn-x")
	if err != nil || logID == uuid.Nil {
		t.Fatalf("record: logID=%v err=%v", logID, err)
	}
}

func TestAutoLogoutWithForceLogoutAndDisconnect(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	force := testutil.NewMemForceLogout()
	disconnected := false
	forceDisconnect := func(_ context.Context, _ uuid.UUID) error {
		disconnected = true
		return nil
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	login, err := testutil.NewAuthUC(t, repos).Login(context.Background(), usecase.LoginInput{
		Email: "alice@test.local", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	old := time.Now().UTC().Add(-2 * time.Hour)
	user.LastActivityAt = old
	user.UpdatedAt = old
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("update activity: %v", err)
	}

	auto := usecase.NewAutoLogoutUseCase(repos, force, forceDisconnect, time.Hour)
	if err := auto.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !disconnected {
		t.Fatalf("force disconnect should be called")
	}
	before, err := force.GetForceLogoutBefore(context.Background(), user.ID)
	if err != nil || before.IsZero() {
		t.Fatalf("force logout: %v", before)
	}

	_, err = testutil.NewAuthUC(t, repos).Refresh(context.Background(), usecase.RefreshInput{
		RefreshToken: login.RefreshToken,
	})
	if err == nil {
		t.Fatalf("refresh should fail after auto logout")
	}
}

func TestAutoLogoutCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auto := usecase.NewAutoLogoutUseCase(repos, nil, nil, time.Hour)
	auto.Now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	if err := auto.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestRankingRebuildExcludesDeletedUsers(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	active := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	deleted := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	active.WinCount = 5
	deleted.WinCount = 10
	now := time.Now().UTC()
	deleted.DeletedAt = &now
	ctx := context.Background()
	if err := repos.User.Update(ctx, active); err != nil {
		t.Fatalf("update active: %v", err)
	}
	if err := repos.User.Update(ctx, deleted); err != nil {
		t.Fatalf("update deleted: %v", err)
	}

	if err := admin.RebuildRanking(ctx, master.ID); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	rows, err := ranking.Get(ctx)
	if err != nil || len(rows) != 1 || rows[0].Username != "alice" {
		t.Fatalf("ranking excludes deleted: %+v err=%v", rows, err)
	}
}

func TestRunScheduledRebuildNilLocks(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := usecase.NewRankingUseCase(repos, nil, 0)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	a.WinCount = 2
	if err := repos.User.Update(context.Background(), a); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := ranking.RunScheduledRebuild(context.Background()); err != nil {
		t.Fatalf("scheduled rebuild nil locks: %v", err)
	}
}

func TestRankingCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	ranking.Now = func() time.Time { return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC) }
	if err := ranking.Rebuild(context.Background()); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
}

func TestLogRetentionMultiBatchPurge(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	for i := 0; i < 5; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "batch_old", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
		}); err != nil {
			t.Fatalf("activity log: %v", err)
		}
	}

	retention := usecase.NewLogRetentionUseCase(repos, 30, 0, 0, 2, time.Millisecond)
	retention.Run(context.Background())

	rows, _, err := repos.ActivityLog.Search(context.Background(), "batch_old", nil, nil, nil, 1, 10)
	if err != nil || len(rows) != 0 {
		t.Fatalf("batch purge incomplete: %+v err=%v", rows, err)
	}
}

func TestLogRetentionCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	retention := usecase.NewLogRetentionUseCase(repos, 30, 0, 0, 100, 0)
	retention.Now = func() time.Time { return time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC) }
	retention.Run(context.Background())
}

func TestLogRetentionCancelDuringBatchSleep(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	for i := 0; i < 4; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "sleep_old", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
		}); err != nil {
			t.Fatalf("activity log: %v", err)
		}
	}

	retention := usecase.NewLogRetentionUseCase(repos, 30, 0, 0, 1, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	retention.Run(ctx)
}

func TestBackupCustomNow(t *testing.T) {
	backupUC := usecase.NewBackupUseCase(nil, nil, 0)
	backupUC.Now = func() time.Time { return time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC) }
	if err := backupUC.RunSync(context.Background()); err != nil {
		t.Fatalf("nil syncer: %v", err)
	}
}

func TestGetProfileNoRankingEntry(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	out, err := profile.GetProfile(context.Background(), user.ID)
	if err != nil || out.Rank != nil {
		t.Fatalf("profile without rank: %+v err=%v", out, err)
	}
}

func TestGetProfileRepoError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	profile := usecase.NewProfileUseCase(repos)

	_, err := profile.GetProfile(context.Background(), uuid.New())
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("missing user profile: %v", err)
	}
}
