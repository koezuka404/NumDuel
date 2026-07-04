package usecase

import (
	"context"
	"encoding/csv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// セキュリティ: CSV formula injection 対策
func TestSanitizeCSVCell(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"=1+1", "'=1+1"},
		{"+cmd", "'+cmd"},
		{"-2+3", "'-2+3"},
		{"@SUM(A1)", "'@SUM(A1)"},
		{"safe", "safe"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := sanitizeCSVCell(tt.in); got != tt.want {
			t.Fatalf("sanitizeCSVCell(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

type testAdminLockStore struct{}

func (testAdminLockStore) AcquireLock(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}

func TestDownloadCSVWriterFlushError(t *testing.T) {
	repos := openTestRepos(t)
	orig := flushCSVWriter
	flushCSVWriter = func(*csv.Writer) error { return context.Canceled }
	t.Cleanup(func() { flushCSVWriter = orig })

	admin := NewAdminUseCase(repos, nil, nil, nil, nil, testAdminLockStore{}, time.Minute)
	now := time.Now().UTC()
	master := &model.User{
		ID: uuid.New(), Username: "admin", Email: "admin@test.local", PasswordHash: "h",
		Role: model.RoleMaster, LastActivityAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.User.Create(context.Background(), master); err != nil {
		t.Fatalf("create master: %v", err)
	}

	_, err := admin.DownloadActivityLogsCSV(context.Background(), master.ID, "", nil, nil, nil)
	if err == nil {
		t.Fatal("expected csv flush error")
	}
}
