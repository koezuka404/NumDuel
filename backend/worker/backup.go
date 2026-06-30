package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/numduel/numduel/usecase"
)

// BackupWorker は BACKUP_CRON に従い primary → backup DB へ差分 UPSERT する（§12.8）
type BackupWorker struct {
	Deps usecase.BackupDeps
	Cron string
}

func (w *BackupWorker) Run(ctx context.Context) {
	if w.Deps.Syncer == nil || w.Cron == "" {
		return
	}
	sched := cron.New(cron.WithLocation(time.UTC))
	if _, err := sched.AddFunc(w.Cron, func() {
		if err := usecase.RunBackupSync(ctx, w.Deps); err != nil {
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
