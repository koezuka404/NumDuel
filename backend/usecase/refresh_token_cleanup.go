package usecase

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/repository"
)

const defaultRefreshTokenCleanupGraceDays = 7

// RefreshTokenCleanupDeps は RefreshTokenCleanupWorker / RunRefreshTokenCleanup の依存
type RefreshTokenCleanupDeps struct {
	Repo      repository.Repos
	GraceDays int
	Now       func() time.Time
}

// RunRefreshTokenCleanup は猶予期間を過ぎた refresh_tokens を物理削除する（§12.1 / §13.9.4）
func RunRefreshTokenCleanup(ctx context.Context, d RefreshTokenCleanupDeps) {
	if d.Repo == nil {
		return
	}
	graceDays := d.GraceDays
	if graceDays <= 0 {
		graceDays = defaultRefreshTokenCleanupGraceDays
	}
	cutoff := d.now().AddDate(0, 0, -graceDays)

	n, err := d.Repo.RefreshToken.DeleteExpired(ctx, cutoff)
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("refresh token cleanup: %v", err)
		}
		return
	}
	if n > 0 {
		log.Printf("refresh token cleanup: deleted %d rows", n)
	}
}

func (d RefreshTokenCleanupDeps) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}
