package repository

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestMapRecordNotFound(t *testing.T) {
	if mapRecordNotFound(gorm.ErrRecordNotFound) != ErrNotFound {
		t.Fatal("expected ErrNotFound")
	}
	other := errors.New("db down")
	if mapRecordNotFound(other) != other {
		t.Fatal("expected original error")
	}
}

func TestFirstOne(t *testing.T) {
	db := openTestDB(t)
	now := mustUser(t, db)

	var found model.User
	if err := firstOne(db.Where("id = ?", now.ID), &found); err != nil {
		t.Fatalf("firstOne found: %v", err)
	}
	if found.ID != now.ID {
		t.Fatalf("id mismatch")
	}

	var missing model.User
	if err := firstOne(db.Where("id = ?", uuid.New()), &missing); !errors.Is(err, ErrNotFound) {
		t.Fatalf("firstOne missing: %v", err)
	}
}

func TestFirstOneForUpdate(t *testing.T) {
	db := openTestDB(t)
	now := mustUser(t, db)

	var found model.User
	if err := firstOneForUpdate(db.Where("id = ?", now.ID), &found); err != nil {
		t.Fatalf("firstOneForUpdate found: %v", err)
	}

	var missing model.User
	if err := firstOneForUpdate(db.Where("id = ?", uuid.New()), &missing); !errors.Is(err, ErrNotFound) {
		t.Fatalf("firstOneForUpdate missing: %v", err)
	}
}

func TestFindOptional(t *testing.T) {
	db := openTestDB(t)
	now := mustUser(t, db)

	got, err := findOptional[model.User](db.Where("id = ?", now.ID))
	if err != nil || got == nil || got.ID != now.ID {
		t.Fatalf("findOptional found: %+v err=%v", got, err)
	}

	got, err = findOptional[model.User](db.Where("id = ?", uuid.New()))
	if err != nil || got != nil {
		t.Fatalf("findOptional missing: %+v err=%v", got, err)
	}
}

func TestFindOptionalForUpdate(t *testing.T) {
	db := openTestDB(t)
	now := mustUser(t, db)

	got, err := findOptionalForUpdate[model.User](db.Where("id = ?", now.ID))
	if err != nil || got == nil || got.ID != now.ID {
		t.Fatalf("findOptionalForUpdate found: %+v err=%v", got, err)
	}

	got, err = findOptionalForUpdate[model.User](db.Where("id = ?", uuid.New()))
	if err != nil || got != nil {
		t.Fatalf("findOptionalForUpdate missing: %+v err=%v", got, err)
	}
}

func TestPaginatePage(t *testing.T) {
	limit, offset := paginatePage(0, 0)
	if limit != 20 || offset != 0 {
		t.Fatalf("defaults: limit=%d offset=%d", limit, offset)
	}
	limit, offset = paginatePage(2, 10)
	if limit != 10 || offset != 10 {
		t.Fatalf("page2: limit=%d offset=%d", limit, offset)
	}
}

func TestRowsAffected(t *testing.T) {
	if err := rowsAffected(nil, 1); err != nil {
		t.Fatalf("rowsAffected ok: %v", err)
	}
	if !errors.Is(rowsAffected(nil, 0), ErrNotFound) {
		t.Fatal("expected ErrNotFound for zero rows")
	}
	if rowsAffected(errors.New("fail"), 1) == nil {
		t.Fatal("expected error passthrough")
	}
}

func mustUser(t *testing.T, db *gorm.DB) model.User {
	t.Helper()
	user := model.User{
		ID: uuid.New(), Username: "alice", Email: "alice@test.local",
		PasswordHash: "hash", Role: model.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}
