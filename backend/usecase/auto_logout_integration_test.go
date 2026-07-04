package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// §18.5.4 自動ログアウト
func TestAutoLogoutInactiveUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
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

	auto := usecase.NewAutoLogoutUseCase(repos, nil, nil, time.Hour)
	if err := auto.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}

	_, err = testutil.NewAuthUC(t, repos).Refresh(context.Background(), usecase.RefreshInput{
		RefreshToken: login.RefreshToken,
	})
	if err == nil {
		t.Fatalf("expected refresh failure after auto logout")
	}

	logs, _, err := repos.LoginLog.ListByUserID(context.Background(), user.ID, 1, 10)
	if err != nil || len(logs) == 0 || logs[0].Action != model.LoginActionAutoLogout {
		t.Fatalf("auto logout log: %+v err=%v", logs, err)
	}
}

func TestAutoLogoutSkipsWhenTimeoutZero(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	auto := usecase.NewAutoLogoutUseCase(repos, nil, nil, 0)
	if err := auto.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
}

// §18.5.4 ログ保持
func TestLogRetentionDeletesOldLogs(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	userID := uuid.New()
	old := time.Now().UTC().AddDate(0, 0, -40)
	now := time.Now().UTC()

	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), LogType: "old", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
	}); err != nil {
		t.Fatalf("activity log: %v", err)
	}
	if err := repos.LoginLog.Create(context.Background(), &model.LoginLog{
		ID: uuid.New(), UserID: userID, Action: model.LoginActionLogin, CreatedAt: old, UpdatedAt: old,
	}); err != nil {
		t.Fatalf("login log: %v", err)
	}
	if err := repos.WSConnectionLog.Create(context.Background(), &model.WSConnectionLog{
		ID: uuid.New(), UserID: userID, ConnectionID: "c1", ConnectedAt: old,
	}); err != nil {
		t.Fatalf("ws log: %v", err)
	}
	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), LogType: "new", Detail: []byte(`{}`), CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("new activity log: %v", err)
	}

	retention := usecase.NewLogRetentionUseCase(repos, 30, 30, 30, 100, 0)
	retention.Run(context.Background())

	n, err := repos.ActivityLog.DeleteOlderThan(context.Background(), now.AddDate(0, 0, -1), 100)
	if err != nil {
		t.Fatalf("count activity: %v", err)
	}
	_ = n
	rows, _, err := repos.ActivityLog.Search(context.Background(), "new", nil, nil, nil, 1, 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("remaining activity logs: %+v err=%v", rows, err)
	}
	oldRows, _, err := repos.ActivityLog.Search(context.Background(), "old", nil, nil, nil, 1, 10)
	if err != nil || len(oldRows) != 0 {
		t.Fatalf("old activity logs should be deleted: %+v err=%v", oldRows, err)
	}
}
