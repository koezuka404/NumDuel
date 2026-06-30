package worker

import (
	"context"
	"log"
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type SecretSetupTimeoutWorker struct {
	Game *usecase.GameUseCase
	Interval time.Duration
}

func (w *SecretSetupTimeoutWorker) Run(ctx context.Context) {
	if w.Game == nil || w.Interval <= 0 {
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
	games, err := w.Game.Games.ListByStatusCreatedBefore(ctx, model.GameStatusWaitingSecret, before)
	if err != nil {
		log.Printf("secret setup timeout worker: list games: %v", err)
		return
	}
	for _, game := range games {
		if game == nil {
			continue
		}
		if err := w.Game.CancelBySecretTimeout(ctx, game.ID); err != nil {
			log.Printf("secret setup timeout worker: game=%s: %v", game.ID, err)
		}
	}
}
