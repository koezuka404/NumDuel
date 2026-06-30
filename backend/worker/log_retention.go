package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

type LogRetentionWorker struct {
	Retention *usecase.LogRetentionUseCase
	Cron      string
}

func (w *LogRetentionWorker) Run(ctx context.Context) {
	if w.Retention == nil || w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		w.Retention.Run(ctx)
	}); err != nil {
		log.Printf("log retention worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
