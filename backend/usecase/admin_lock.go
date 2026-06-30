package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// rankingRebuildWorkerActorID は RankingRebuildWorker 用 admin ロックの actor ID
var rankingRebuildWorkerActorID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func adminRankingRebuildLockKey(adminID uuid.UUID) string {
	return fmt.Sprintf("admin:%s:ranking_rebuild_lock", adminID)
}

func adminLogDownloadLockKey(adminID uuid.UUID) string {
	return fmt.Sprintf("admin:%s:log_download_lock", adminID)
}

func adminUserDeleteLockKey(adminID uuid.UUID) string {
	return fmt.Sprintf("admin:%s:user_delete_lock", adminID)
}

func acquireAdminLock(ctx context.Context, d AdminDeps, key string) error {
	if d.Locks == nil {
		return nil
	}
	ttl := d.AdminLockTTL
	if ttl <= 0 {
		ttl = 5 * time.Second
	}
	ok, err := d.Locks.AcquireLock(ctx, key, ttl)
	if err != nil {
		return model.ErrInternal("failed to acquire admin lock")
	}
	if !ok {
		return model.ErrRateLimitExceeded()
	}
	return nil
}

func acquireRankingRebuildLock(ctx context.Context, locks model.GameLockStore, actorID uuid.UUID, ttl time.Duration) (bool, error) {
	if locks == nil {
		return true, nil
	}
	if ttl <= 0 {
		ttl = 5 * time.Second
	}
	return locks.AcquireLock(ctx, adminRankingRebuildLockKey(actorID), ttl)
}
