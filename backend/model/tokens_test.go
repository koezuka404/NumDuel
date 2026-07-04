package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestRefreshTokenHelpers(t *testing.T) {
	now := time.Now().UTC()
	active := model.NewRefreshToken(uuid.New(), "hash", uuid.New(), now.Add(time.Hour), now)
	if !active.IsActive(now) || active.IsRevoked() || active.IsExpired(now) {
		t.Fatal("active token")
	}
	expired := model.NewRefreshToken(uuid.New(), "h2", uuid.New(), now.Add(-time.Hour), now)
	if !expired.IsExpired(now) || expired.IsActive(now) {
		t.Fatal("expired token")
	}
	revoked := active
	revoked.Status = model.RefreshTokenRevoked
	if !revoked.IsRevoked() || revoked.IsActive(now) {
		t.Fatal("revoked token")
	}
	if (model.RefreshToken{}).TableName() != "refresh_tokens" {
		t.Fatal("TableName")
	}
}
