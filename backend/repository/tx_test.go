package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

func TestWithTxCommitAndRollback(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	ctx := context.Background()
	user := newUser("txuser", "tx@test.local")

	err := repository.WithTx(ctx, gdb, func(txCtx context.Context) error {
		return repos.User.Create(txCtx, user)
	})
	if err != nil {
		t.Fatalf("commit tx: %v", err)
	}
	got, err := repos.User.FindByID(ctx, user.ID)
	if err != nil || got == nil {
		t.Fatalf("user persisted: %+v err=%v", got, err)
	}

	user2 := newUser("rollback", "rollback@test.local")
	err = repository.WithTx(ctx, gdb, func(txCtx context.Context) error {
		if err := repos.User.Create(txCtx, user2); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}
	got, err = repos.User.FindByID(ctx, user2.ID)
	if err != nil || got != nil {
		t.Fatalf("user rolled back: %+v err=%v", got, err)
	}
}
