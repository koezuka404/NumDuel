package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

// RefreshTokenCleanupWorker は REFRESH_TOKEN_CLEANUP_CRON に従い refresh_tokens を削除する
type RefreshTokenCleanupWorker struct {
	Deps usecase.RefreshTokenCleanupDeps
	Cron string
}

func (w *RefreshTokenCleanupWorker) Run(ctx context.Context) {
	if w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		usecase.RunRefreshTokenCleanup(ctx, w.Deps)
	}); err != nil {
		log.Printf("refresh token cleanup worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
