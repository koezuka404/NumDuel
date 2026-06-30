package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

type RefreshTokenCleanupWorker struct {
	Auth *usecase.AuthUseCase
	Cron string
}

func (w *RefreshTokenCleanupWorker) Run(ctx context.Context) {
	if w.Cron == "" || w.Auth == nil {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		w.Auth.CleanupExpiredRefreshTokens(ctx)
	}); err != nil {
		log.Printf("refresh token cleanup worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
