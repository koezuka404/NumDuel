package model_test

import (
	"testing"

	"github.com/numduel/numduel/model"
)

func TestEnumValid(t *testing.T) {
	if !model.GameStatusInProgress.Valid() || model.GameStatus("bad").Valid() {
		t.Fatal("GameStatus Valid")
	}
	if !model.RoleUser.Valid() || !model.RoleMaster.Valid() || model.Role("x").Valid() {
		t.Fatal("Role Valid")
	}
	if !model.DigitHit.Valid() || !model.DigitMiss.Valid() || model.DigitResult(2).Valid() {
		t.Fatal("DigitResult Valid")
	}
	if !model.RefreshTokenActive.Valid() || model.RefreshTokenStatus("x").Valid() {
		t.Fatal("RefreshTokenStatus Valid")
	}
	if !model.MatchingQueueWaiting.Valid() || model.MatchingQueueStatus("x").Valid() {
		t.Fatal("MatchingQueueStatus Valid")
	}
	if !model.LoginActionLogin.Valid() || model.LoginAction("x").Valid() {
		t.Fatal("LoginAction Valid")
	}
}
