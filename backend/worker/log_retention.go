package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

// LogRetentionWorker は LOG_RETENTION_CRON に従い古いログをバッチ削除する（§12.7）
type LogRetentionWorker struct {
	Deps usecase.LogRetentionDeps
	Cron string
}

func (w *LogRetentionWorker) Run(ctx context.Context) {
	if w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		usecase.RunLogRetention(ctx, w.Deps)
	}); err != nil {
		log.Printf("log retention worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
