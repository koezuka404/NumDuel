package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

type RankingRebuildWorker struct {
	Ranking *usecase.RankingUseCase
	Cron    string
}

func (w *RankingRebuildWorker) Run(ctx context.Context) {
	if w.Ranking == nil || w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		if err := w.Ranking.RunScheduledRebuild(ctx); err != nil {
			log.Printf("ranking rebuild worker: %v", err)
		}
	}); err != nil {
		log.Printf("ranking rebuild worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
