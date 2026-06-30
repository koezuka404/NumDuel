package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

func (a *AdminUseCase) acquireLock(ctx context.Context, key string) error {
	if a == nil || a.Locks == nil {
		return nil
	}
	ttl := a.AdminLockTTL
	if ttl <= 0 {
		ttl = 5 * time.Second
	}
	ok, err := a.Locks.AcquireLock(ctx, key, ttl)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRateLimitExceeded
	}
	return nil
}

func acquireRankingRebuildLock(ctx context.Context, locks IDistributedLockStore, actorID uuid.UUID, ttl time.Duration) (bool, error) {
	if locks == nil {
		return true, nil
	}
	if ttl <= 0 {
		ttl = 5 * time.Second
	}
	return locks.AcquireLock(ctx, adminRankingRebuildLockKey(actorID), ttl)
}
