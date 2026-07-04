package dto

import (
	"errors"
	"testing"

	"github.com/numduel/numduel/usecase"
)

func TestConflictCodeDefault(t *testing.T) {
	if conflictCode(errors.New("unknown")) != "conflict" {
		t.Fatal("default conflict code")
	}
	if conflictCode(usecase.ErrDuplicateUser) != "duplicate_user" {
		t.Fatal("duplicate_user")
	}
}
