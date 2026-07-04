package repository

import (
	"testing"

	"github.com/numduel/numduel/model"
)

func TestFindOptionalGenericError(t *testing.T) {
	db := openTestDB(t)
	got, err := findOptional[model.User](db.Table("not_a_real_table"))
	if err == nil || got != nil {
		t.Fatalf("expected generic error, got %+v err=%v", got, err)
	}
}

func TestFindOptionalForUpdateGenericError(t *testing.T) {
	db := openTestDB(t)
	got, err := findOptionalForUpdate[model.User](db.Table("not_a_real_table"))
	if err == nil || got != nil {
		t.Fatalf("expected generic error, got %+v err=%v", got, err)
	}
}
