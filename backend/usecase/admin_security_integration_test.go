package usecase_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
)

func TestAdminCSVExportSanitizesFormulaCells(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	now := time.Now().UTC()

	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), LogType: "security_test",
		Detail: json.RawMessage("=1+1"), CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create log: %v", err)
	}

	csv, err := admin.DownloadActivityLogsCSV(context.Background(), master.ID, "security_test", nil, nil, nil)
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	if !strings.Contains(string(csv), "'=1+1") {
		t.Fatalf("expected sanitized csv, got:\n%s", string(csv))
	}
}
