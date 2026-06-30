package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// 分散ロック（ランキング再構築・管理操作）。
type IDistributedLockStore interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// ランキング取得・再構築ユースケース。
type IRankingUsecase interface {
	Get(ctx context.Context) ([]RankingItem, error)
	Rebuild(ctx context.Context) error
	RunScheduledRebuild(ctx context.Context) error
}

type RankingUseCase struct {
	Rankings repository.IRankingRepo
	Repos    repository.Repos
	Locks    IDistributedLockStore
	LockTTL  time.Duration
	Now      func() time.Time
}

func (r *RankingUseCase) now() time.Time {
	if r != nil && r.Now != nil {
		return r.Now().UTC()
	}
	return time.Now().UTC()
}

type RankingItem struct {
	Rank     int
	Username string
	WinCount int
}

func (r *RankingUseCase) Get(ctx context.Context) ([]RankingItem, error) {
	rows, err := r.Rankings.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) > 3 {
		rows = rows[:3]
	}
	out := make([]RankingItem, len(rows))
	for i, row := range rows {
		out[i] = RankingItem{Rank: row.Rank, Username: row.Username, WinCount: row.WinCount}
	}
	return out, nil
}

func (r *RankingUseCase) Rebuild(ctx context.Context) error {
	rows, err := listUsersForRankingRebuild(ctx, r.Repos)
	if err != nil {
		return err
	}
	now := r.now()
	rankings := make([]model.Ranking, len(rows))
	for i, row := range rows {
		rankings[i] = model.Ranking{
			UserID: row.UserID, Rank: i + 1, Username: row.Username, WinCount: row.WinCount, UpdatedAt: now,
		}
	}
	return repository.WithTx(ctx, r.Repos.DB, func(ctx context.Context) error {
		return r.Rankings.ReplaceAll(ctx, rankings)
	})
}

func (r *RankingUseCase) RunScheduledRebuild(ctx context.Context) error {
	ok, err := acquireRankingRebuildLock(ctx, r.Locks, rankingRebuildWorkerActorID, r.LockTTL)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return r.Rebuild(ctx)
}

type rankingRebuildRow struct {
	UserID   uuid.UUID
	Username string
	WinCount int
}

func listUsersForRankingRebuild(ctx context.Context, repo repository.Repos) ([]rankingRebuildRow, error) {
	users, err := repo.User.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]rankingRebuildRow, 0, len(users))
	for _, u := range users {
		if u.IsDeleted() || u.IsMaster() {
			continue
		}
		rows = append(rows, rankingRebuildRow{
			UserID:   u.ID,
			Username: u.Username,
			WinCount: u.WinCount,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].WinCount != rows[j].WinCount {
			return rows[i].WinCount > rows[j].WinCount
		}
		return rows[i].Username < rows[j].Username
	})
	return rows, nil
}

func NewRankingUseCase(repos repository.Repos, locks IDistributedLockStore, lockTTL time.Duration) *RankingUseCase {
	return &RankingUseCase{
		Rankings: repos.Ranking,
		Repos:    repos,
		Locks:    locks,
		LockTTL:  lockTTL,
	}
}
