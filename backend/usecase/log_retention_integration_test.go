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

func TestLogRetentionSkipsZeroDays(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), LogType: "old", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
	}); err != nil {
		t.Fatalf("activity log: %v", err)
	}

	retention := usecase.NewLogRetentionUseCase(repos, 0, 0, 0, 100, 0)
	retention.Run(context.Background())

	rows, _, err := repos.ActivityLog.Search(context.Background(), "old", nil, nil, nil, 1, 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("log should remain when days=0: %+v err=%v", rows, err)
	}
}

func TestLogRetentionContextCanceled(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	retention := usecase.NewLogRetentionUseCase(repos, 30, 30, 30, 100, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	retention.Run(ctx)
}

func TestLogRetentionUsesDefaultBatchSize(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	old := time.Now().UTC().AddDate(0, 0, -40)
	for i := 0; i < 3; i++ {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: "old", Detail: []byte(`{}`), CreatedAt: old, UpdatedAt: old,
		}); err != nil {
			t.Fatalf("activity log: %v", err)
		}
	}

	retention := usecase.NewLogRetentionUseCase(repos, 30, 0, 0, 0, 0)
	retention.Run(context.Background())

	rows, _, err := repos.ActivityLog.Search(context.Background(), "old", nil, nil, nil, 1, 10)
	if err != nil || len(rows) != 0 {
		t.Fatalf("old logs should be purged: %+v err=%v", rows, err)
	}
}
