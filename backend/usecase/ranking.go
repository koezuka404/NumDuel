package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type RankingDeps struct {
	Repo repository.IRepository
	Tx   repository.ITxManager
	Now  func() time.Time
}

func (d RankingDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

type RankingItem struct {
	Rank     int
	Username string
	WinCount int
}

// GetRanking は上位 3 名を返す
func GetRanking(ctx context.Context, d RankingDeps) ([]RankingItem, error) {
	rows, err := d.Repo.Rankings().ListAll(ctx)
	if err != nil {
		return nil, model.ErrInternal("failed to load rankings")
	}
	if len(rows) > 3 {
		rows = rows[:3]
	}
	out := make([]RankingItem, len(rows))
	for i, r := range rows {
		out[i] = RankingItem{Rank: r.Rank, Username: r.Username, WinCount: r.WinCount}
	}
	return out, nil
}

// RebuildRanking は users.win_count から rankings を全件再集計する
func RebuildRanking(ctx context.Context, d RankingDeps) error {
	rows, err := listUsersForRankingRebuild(ctx, d.Repo)
	if err != nil {
		return model.ErrInternal("failed to load users for ranking")
	}
	now := d.now()
	rankings := make([]model.Ranking, len(rows))
	for i, row := range rows {
		rankings[i] = model.NewRanking(row.UserID, i+1, row.Username, row.WinCount, now)
	}
	return d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		if err := tx.Rankings().ReplaceAll(ctx, rankings); err != nil {
			return model.ErrInternal("failed to replace rankings")
		}
		return nil
	})
}

// RankingRebuildWorkerDeps は RankingRebuildWorker の依存
type RankingRebuildWorkerDeps struct {
	Ranking RankingDeps
	Locks   model.IGameLockStore
	LockTTL time.Duration
}

// RunScheduledRankingRebuild は cron から rankings を再集計する（§12.6）
func RunScheduledRankingRebuild(ctx context.Context, d RankingRebuildWorkerDeps) error {
	ok, err := acquireRankingRebuildLock(ctx, d.Locks, rankingRebuildWorkerActorID, d.LockTTL)
	if err != nil {
		return model.ErrInternal("failed to acquire ranking rebuild lock")
	}
	if !ok {
		return nil
	}
	return RebuildRanking(ctx, d.Ranking)
}
