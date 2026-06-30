package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

// RankingRebuildWorker は RANKING_REBUILD_CRON に従い rankings を再集計する（§12.6）
type RankingRebuildWorker struct {
	Deps usecase.RankingRebuildWorkerDeps
	Cron string
}

func (w *RankingRebuildWorker) Run(ctx context.Context) {
	if w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		if err := usecase.RunScheduledRankingRebuild(ctx, w.Deps); err != nil {
			if de, ok := model.IsDomainError(err); ok && de.Code == model.CodeInternalError {
				log.Printf("ranking rebuild worker: %v", err)
			}
		}
	}); err != nil {
		log.Printf("ranking rebuild worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
