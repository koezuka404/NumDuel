package worker

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/usecase"
)

type AutoLogoutWorker struct {
	AutoLogout *usecase.AutoLogoutUseCase
	Interval   time.Duration
}

func (w *AutoLogoutWorker) Run(ctx context.Context) {
	if w.AutoLogout == nil || w.Interval <= 0 {
		return
	}
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			w.tick(ctx, t.UTC())
		}
	}
}

func (w *AutoLogoutWorker) tick(ctx context.Context, now time.Time) {
	uc := w.AutoLogout
	if uc.Now == nil {
		uc.Now = func() time.Time { return now }
	}
	if err := uc.Run(ctx); err != nil {
		log.Printf("auto logout worker: %v", err)
	}
}
