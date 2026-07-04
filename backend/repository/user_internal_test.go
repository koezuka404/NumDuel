package repository

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

func TestUserSearchScopePostgresBranch(t *testing.T) {
	orig := isPostgresDialect
	isPostgresDialect = func(string) bool { return true }
	t.Cleanup(func() { isPostgresDialect = orig })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	var users []model.User
	stmt := userSearchScope(db, "%alice%").Find(&users).Statement
	sql := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
	if !strings.Contains(sql, "ILIKE") {
		t.Fatalf("expected ILIKE in postgres search SQL, got: %s", sql)
	}
}
