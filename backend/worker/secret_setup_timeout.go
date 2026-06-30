package worker

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

// SecretSetupTimeoutWorker は WAITING_SECRET の期限切れゲームをポーリングして終了する
type SecretSetupTimeoutWorker struct {
	Game     usecase.GameDeps
	Interval time.Duration
}

func (w *SecretSetupTimeoutWorker) Run(ctx context.Context) {
	if w.Interval <= 0 {
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

func (w *SecretSetupTimeoutWorker) tick(ctx context.Context, now time.Time) {
	if w.Game.SecretSetup <= 0 {
		return
	}
	before := now.Add(-w.Game.SecretSetup)
	games, err := w.Game.Repo.Game.ListByStatusCreatedBefore(ctx, model.GameStatusWaitingSecret, before)
	if err != nil {
		log.Printf("secret setup timeout worker: list games: %v", err)
		return
	}
	for _, game := range games {
		if game == nil {
			continue
		}
		if err := usecase.CancelGameBySecretTimeout(ctx, w.Game, game.ID); err != nil {
			if de, ok := model.IsDomainError(err); ok && de.Code == model.CodeInternalError {
				log.Printf("secret setup timeout worker: game=%s: %v", game.ID, err)
			}
		}
	}
}
