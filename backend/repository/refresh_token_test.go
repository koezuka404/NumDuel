package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func TestRefreshTokenRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	user := createUser(t, repos, "alice", "alice@test.local")

	now := time.Now().UTC()
	familyID := uuid.New()
	token := model.NewRefreshToken(user.ID, "hash-active", familyID, now.Add(24*time.Hour), now)
	if err := repos.RefreshToken.Create(ctx, &token); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repos.RefreshToken.FindByTokenHash(ctx, "hash-active")
	if err != nil || got == nil || got.ID != token.ID {
		t.Fatalf("find by hash: %+v err=%v", got, err)
	}

	got, err = repos.RefreshToken.FindByTokenHashForUpdate(ctx, "hash-active")
	if err != nil || got == nil {
		t.Fatalf("find for update: %+v err=%v", got, err)
	}

	replacementID := uuid.New()
	if err := repos.RefreshToken.MarkUsed(ctx, token.ID, now, replacementID); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	if err := repos.RefreshToken.MarkUsed(ctx, token.ID, now, replacementID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("mark used again: %v", err)
	}

	active2 := model.NewRefreshToken(user.ID, "hash-revoke", familyID, now.Add(24*time.Hour), now)
	if err := repos.RefreshToken.Create(ctx, &active2); err != nil {
		t.Fatalf("create active2: %v", err)
	}
	if err := repos.RefreshToken.Revoke(ctx, active2.ID, now); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := repos.RefreshToken.Revoke(ctx, active2.ID, now); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("revoke again: %v", err)
	}

	active3 := model.NewRefreshToken(user.ID, "hash-family", familyID, now.Add(24*time.Hour), now)
	if err := repos.RefreshToken.Create(ctx, &active3); err != nil {
		t.Fatalf("create active3: %v", err)
	}
	if err := repos.RefreshToken.RevokeByFamilyID(ctx, familyID, now); err != nil {
		t.Fatalf("revoke family: %v", err)
	}

	active4 := model.NewRefreshToken(user.ID, "hash-user", uuid.New(), now.Add(24*time.Hour), now)
	if err := repos.RefreshToken.Create(ctx, &active4); err != nil {
		t.Fatalf("create active4: %v", err)
	}
	if err := repos.RefreshToken.RevokeByUserID(ctx, user.ID, now); err != nil {
		t.Fatalf("revoke user: %v", err)
	}

	expired := model.NewRefreshToken(user.ID, "hash-expired", uuid.New(), now.Add(-time.Hour), now.Add(-2*time.Hour))
	if err := repos.RefreshToken.Create(ctx, &expired); err != nil {
		t.Fatalf("create expired: %v", err)
	}
	oldRevoked := model.NewRefreshToken(user.ID, "hash-old-revoked", uuid.New(), now.Add(24*time.Hour), now.Add(-48*time.Hour))
	revokedAt := now.Add(-49 * time.Hour)
	oldRevoked.Status = model.RefreshTokenRevoked
	oldRevoked.RevokedAt = &revokedAt
	if err := repos.RefreshToken.Create(ctx, &oldRevoked); err != nil {
		t.Fatalf("create old revoked: %v", err)
	}
	n, err := repos.RefreshToken.DeleteExpired(ctx, now.Add(-24*time.Hour))
	if err != nil || n == 0 {
		t.Fatalf("delete expired: n=%d err=%v", n, err)
	}
}
