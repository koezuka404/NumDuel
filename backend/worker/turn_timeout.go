package worker

import (
	"context"
	"log"
	"time"

	infrredis "github.com/numduel/numduel/redis"
	"github.com/numduel/numduel/usecase"
)

type TurnTimeoutWorker struct {
	Store    *infrredis.Store
	Game *usecase.GameUseCase
	Interval time.Duration
}

func (w *TurnTimeoutWorker) Run(ctx context.Context) {
	if w.Store == nil || w.Game == nil || w.Interval <= 0 {
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

func (w *TurnTimeoutWorker) tick(ctx context.Context, now time.Time) {
	entries, err := w.Store.ListExpiredTurns(ctx, now)
	if err != nil {
		log.Printf("turn timeout worker: list expired turns: %v", err)
		return
	}
	for _, e := range entries {
		if err := w.Game.HandleTimeout(ctx, e.GameID, e.PlayerID); err != nil {
			log.Printf("turn timeout worker: game=%s player=%s: %v", e.GameID, e.PlayerID, err)
		}
	}
}
