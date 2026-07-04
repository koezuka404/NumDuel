package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestMatchingQueueRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	u1 := createUser(t, repos, "alice", "alice@test.local")
	u2 := createUser(t, repos, "bob", "bob@test.local")

	e1 := &model.MatchingQueueEntry{
		ID: uuid.New(), UserID: u1.ID, Status: model.MatchingQueueWaiting, CreatedAt: time.Now().UTC(),
	}
	e2 := &model.MatchingQueueEntry{
		ID: uuid.New(), UserID: u2.ID, Status: model.MatchingQueueWaiting, CreatedAt: time.Now().UTC(),
	}
	if err := repos.MatchingQueue.Insert(ctx, e1); err != nil {
		t.Fatalf("insert e1: %v", err)
	}
	if err := repos.MatchingQueue.Insert(ctx, e2); err != nil {
		t.Fatalf("insert e2: %v", err)
	}

	got, err := repos.MatchingQueue.FindByUserID(ctx, u1.ID)
	if err != nil || got == nil || got.ID != e1.ID {
		t.Fatalf("find by user: %+v err=%v", got, err)
	}

	rows, err := repos.MatchingQueue.ListByStatusForUpdate(ctx, model.MatchingQueueWaiting, 10)
	if err != nil || len(rows) < 2 {
		t.Fatalf("list for update: %d err=%v", len(rows), err)
	}

	if err := repos.MatchingQueue.DeleteByIDs(ctx, nil); err != nil {
		t.Fatalf("delete empty ids: %v", err)
	}
	if err := repos.MatchingQueue.DeleteByIDs(ctx, []uuid.UUID{e1.ID}); err != nil {
		t.Fatalf("delete by ids: %v", err)
	}
	if err := repos.MatchingQueue.DeleteByUserID(ctx, u2.ID); err != nil {
		t.Fatalf("delete by user: %v", err)
	}

	got, err = repos.MatchingQueue.FindByUserID(ctx, u2.ID)
	if err != nil || got != nil {
		t.Fatalf("deleted entry: %+v err=%v", got, err)
	}
}
