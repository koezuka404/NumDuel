package worker

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

// AutoLogoutWorker は無操作ユーザーをポーリングし AutoLogout を実行する
// AUTO_LOGOUT_POLL_SECONDS 間隔で DB を走査、Redis 未接続時は起動しない
type AutoLogoutWorker struct {
	Deps     usecase.AutoLogoutDeps
	Interval time.Duration
}

func (w *AutoLogoutWorker) Run(ctx context.Context) {
	// force_logout_before は Redis 必須
	if w.Deps.ForceLogout == nil || w.Interval <= 0 {
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
	deps := w.Deps
	if deps.Now == nil {
		deps.Now = func() time.Time { return now }
	}
	if err := usecase.AutoLogout(ctx, deps); err != nil {
		if de, ok := model.IsDomainError(err); ok && de.Code == model.CodeInternalError {
			log.Printf("auto logout worker: %v", err)
		}
	}
}
