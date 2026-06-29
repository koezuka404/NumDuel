package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/model"
)

type RankingDeps struct {
	Repo model.Repository
	Tx   model.TxManager
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

// GetRanking は上位 3 名を返す（仕様 6.8.1）。
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

// RebuildRanking は users.win_count から rankings を全件再集計する。
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
	return withTx(ctx, d.Tx, func(tx model.Transaction) error {
		if err := d.Repo.Rankings().ReplaceAll(ctx, tx, rankings); err != nil {
			return model.ErrInternal("failed to replace rankings")
		}
		return nil
	})
}
