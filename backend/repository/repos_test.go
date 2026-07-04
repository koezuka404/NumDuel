package repository_test

import (
	"testing"

	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

func TestNewRepos(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	if repos.DB != gdb {
		t.Fatal("DB mismatch")
	}
	if repos.User == nil || repos.Game == nil || repos.Guess == nil ||
		repos.MatchHistory == nil || repos.MatchingQueue == nil ||
		repos.Ranking == nil || repos.RefreshToken == nil ||
		repos.ActivityLog == nil || repos.LoginLog == nil ||
		repos.WSConnectionLog == nil {
		t.Fatal("expected all repos initialized")
	}
	_ = repository.NewRepos(gdb)
}
