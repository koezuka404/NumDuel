package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

type BackupWorker struct {
	Backup *usecase.BackupUseCase
	Cron   string
}

func (w *BackupWorker) Run(ctx context.Context) {
	if w.Backup == nil || w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		if err := w.Backup.RunSync(ctx); err != nil {
			log.Printf("backup worker: %v", err)
		}
	}); err != nil {
		log.Printf("backup worker: invalid cron %q: %v", w.Cron, err)
		return
	}
	sched.Start()
	defer sched.Stop()
	<-ctx.Done()
}
