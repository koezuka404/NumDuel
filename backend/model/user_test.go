package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestUserHelpers(t *testing.T) {
	u := model.User{Role: model.RoleUser}
	if u.IsDeleted() || u.IsMaster() || !u.CanMatch() {
		t.Fatal("active user")
	}
	now := time.Now().UTC()
	u.DeletedAt = &now
	if !u.IsDeleted() || u.CanMatch() {
		t.Fatal("deleted user")
	}
	m := model.User{Role: model.RoleMaster}
	if !m.IsMaster() || m.CanMatch() {
		t.Fatal("master user")
	}
	if (model.User{}).TableName() != "users" {
		t.Fatal("TableName")
	}
	_ = uuid.New()
}
