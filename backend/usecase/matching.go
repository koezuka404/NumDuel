package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// MatchingDeps はマッチング UseCase の依存関係
type MatchingDeps struct {
	Repo     repository.IRepository
	Tx       repository.TxManager
	Notifier model.EventNotifier
	Now      func() time.Time
}

func (d MatchingDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

type StartMatchingOutput struct {
	Status string // 常に "waiting"（即時ペアリング時も同じ）
}

type CancelMatchingOutput struct {
	Status string // "cancelled"
}

type GetMatchingStatusOutput struct {
	Status string
	GameID *uuid.UUID // matched 時のみ
}

// MatchPlayers は待機キュー先頭 2 人をペアリングして Game を作成する（同一 TX 内）
func MatchPlayers(ctx context.Context, d MatchingDeps, tx repository.ITxRepos) (*model.Game, error) {
	entries, err := tx.MatchingQueue().ListByStatusForUpdate(ctx, model.MatchingQueueWaiting, 2)
	if err != nil {
		return nil, model.ErrInternal("failed to load matching queue")
	}
	if len(entries) < 2 {
		return nil, nil
	}
	p1, p2 := entries[0], entries[1]
	if p1.UserID == p2.UserID {
		return nil, model.ErrInternal("invalid matching pair")
	}
	if ok, err := matchingPlayerReady(ctx, d.Repo, p1.UserID); err != nil {
		return nil, model.ErrInternal("failed to validate matching player")
	} else if !ok {
		return nil, removeQueueEntries(ctx, tx, []uuid.UUID{p1.ID})
	}
	if ok, err := matchingPlayerReady(ctx, d.Repo, p2.UserID); err != nil {
		return nil, model.ErrInternal("failed to validate matching player")
	} else if !ok {
		return nil, removeQueueEntries(ctx, tx, []uuid.UUID{p2.ID})
	}
	now := d.now()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusWaitingSecret,
		Player1ID: p1.UserID, Player2ID: p2.UserID,
		CurrentTurn: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Games().Create(ctx, game); err != nil {
		return nil, model.ErrInternal("failed to create game")
	}
	if err := removeQueueEntries(ctx, tx, []uuid.UUID{p1.ID, p2.ID}); err != nil {
		return nil, err
	}
	return game, nil
}

// StartMatching はキュー登録後、同一 TX 内で MatchPlayers を呼ぶ
func StartMatching(ctx context.Context, d MatchingDeps, userID uuid.UUID) (*StartMatchingOutput, error) {
	user, err := d.Repo.Users().FindByID(ctx, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, model.ErrUnauthorized()
	}
	if user.IsMaster() {
		return nil, model.ErrForbidden("master cannot start matching")
	}
	active, err := FindActiveGameForUser(ctx, d.Repo, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to check active game")
	}
	if active != nil {
		return nil, model.ErrUserInActiveGame()
	}
	waiting, err := userWaitingInMatchingQueue(ctx, d.Repo, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to check matching queue")
	}
	if waiting {
		return nil, model.ErrAlreadyInMatching()
	}
	var matched *model.Game
	now := d.now()
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		if err := tx.MatchingQueue().Insert(ctx, &model.MatchingQueueEntry{
			ID: uuid.New(), UserID: userID, Status: model.MatchingQueueWaiting, CreatedAt: now,
		}); err != nil {
			return model.ErrInternal("failed to join matching queue")
		}
		game, err := MatchPlayers(ctx, d, tx)
		if err != nil {
			return err
		}
		matched = game
		return nil
	}); err != nil {
		return nil, err
	}
	if matched != nil && d.Notifier != nil {
		payload := map[string]any{"gameId": matched.ID.String()}
		_ = d.Notifier.SendToUser(ctx, matched.Player1ID, "MATCHED", payload)
		_ = d.Notifier.SendToUser(ctx, matched.Player2ID, "MATCHED", payload)
	}
	return &StartMatchingOutput{Status: "waiting"}, nil
}

func CancelMatching(ctx context.Context, d MatchingDeps, userID uuid.UUID) (*CancelMatchingOutput, error) {
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		return tx.MatchingQueue().DeleteByUserID(ctx, userID)
	}); err != nil {
		return nil, err
	}
	return &CancelMatchingOutput{Status: "cancelled"}, nil
}

func GetMatchingStatus(ctx context.Context, d MatchingDeps, userID uuid.UUID) (*GetMatchingStatusOutput, error) {
	active, err := FindActiveGameForUser(ctx, d.Repo, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to check active game")
	}
	if active != nil {
		id := active.ID
		return &GetMatchingStatusOutput{Status: "matched", GameID: &id}, nil
	}
	waiting, err := userWaitingInMatchingQueue(ctx, d.Repo, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to check matching queue")
	}
	if waiting {
		return &GetMatchingStatusOutput{Status: "waiting"}, nil
	}
	return &GetMatchingStatusOutput{Status: "idle"}, nil
}

func matchingPlayerReady(ctx context.Context, repo repository.IRepository, userID uuid.UUID) (bool, error) {
	user, err := repo.Users().FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil || user.IsDeleted() || user.IsMaster() {
		return false, nil
	}
	active, err := FindActiveGameForUser(ctx, repo, userID)
	if err != nil {
		return false, err
	}
	return active == nil, nil
}

func removeQueueEntries(ctx context.Context, tx repository.ITxRepos, ids []uuid.UUID) error {
	if err := tx.MatchingQueue().DeleteByIDs(ctx, ids); err != nil {
		return model.ErrInternal("failed to cleanup matching queue")
	}
	return nil
}
