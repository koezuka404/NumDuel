package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// マッチングキューのユースケース。
type IMatchingUsecase interface {
	Start(ctx context.Context, userID uuid.UUID) (*StartMatchingOutput, error)
	Cancel(ctx context.Context, userID uuid.UUID) (*CancelMatchingOutput, error)
	Status(ctx context.Context, userID uuid.UUID) (*GetMatchingStatusOutput, error)
}

type MatchingUseCase struct {
	Users         repository.IUserRepo
	Games         repository.IGameRepo
	MatchingQueue repository.IMatchingQueueRepo
	Repos         repository.Repos
	Notifier      IEventNotifier
	Now           func() time.Time
}

func (m *MatchingUseCase) now() time.Time {
	if m != nil && m.Now != nil {
		return m.Now().UTC()
	}
	return time.Now().UTC()
}

type StartMatchingOutput struct {
	Status string
}

type CancelMatchingOutput struct {
	Status string
}

type GetMatchingStatusOutput struct {
	Status string
	GameID *uuid.UUID
}

func (m *MatchingUseCase) Start(ctx context.Context, userID uuid.UUID) (*StartMatchingOutput, error) {
	user, err := m.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted() {
		return nil, ErrUnauthorized
	}
	if user.IsMaster() {
		return nil, ErrForbidden
	}
	active, err := findActiveGameForUser(ctx, m.Repos, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrUserInActiveGame
	}
	waiting, err := userWaitingInMatchingQueue(ctx, m.Repos, userID)
	if err != nil {
		return nil, err
	}
	if waiting {
		return nil, ErrAlreadyInMatching
	}
	var matched *model.Game
	now := m.now()
	if err := repository.WithTx(ctx, m.Repos.DB, func(ctx context.Context) error {
		if err := m.MatchingQueue.Insert(ctx, &model.MatchingQueueEntry{
			ID: uuid.New(), UserID: userID, Status: model.MatchingQueueWaiting, CreatedAt: now,
		}); err != nil {
			return err
		}
		game, err := m.matchPlayers(ctx)
		if err != nil {
			return err
		}
		matched = game
		return nil
	}); err != nil {
		return nil, err
	}
	if matched != nil && m.Notifier != nil {
		payload := map[string]any{"gameId": matched.ID.String()}
		_ = m.Notifier.SendToUser(ctx, matched.Player1ID, "MATCHED", payload)
		_ = m.Notifier.SendToUser(ctx, matched.Player2ID, "MATCHED", payload)
	}
	return &StartMatchingOutput{Status: "waiting"}, nil
}

func (m *MatchingUseCase) Cancel(ctx context.Context, userID uuid.UUID) (*CancelMatchingOutput, error) {
	if err := repository.WithTx(ctx, m.Repos.DB, func(ctx context.Context) error {
		return m.MatchingQueue.DeleteByUserID(ctx, userID)
	}); err != nil {
		return nil, err
	}
	return &CancelMatchingOutput{Status: "cancelled"}, nil
}

func (m *MatchingUseCase) Status(ctx context.Context, userID uuid.UUID) (*GetMatchingStatusOutput, error) {
	active, err := findActiveGameForUser(ctx, m.Repos, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		id := active.ID
		return &GetMatchingStatusOutput{Status: "matched", GameID: &id}, nil
	}
	waiting, err := userWaitingInMatchingQueue(ctx, m.Repos, userID)
	if err != nil {
		return nil, err
	}
	if waiting {
		return &GetMatchingStatusOutput{Status: "waiting"}, nil
	}
	return &GetMatchingStatusOutput{Status: "idle"}, nil
}

func (m *MatchingUseCase) matchPlayers(ctx context.Context) (*model.Game, error) {
	entries, err := m.MatchingQueue.ListByStatusForUpdate(ctx, model.MatchingQueueWaiting, 2)
	if err != nil {
		return nil, err
	}
	if len(entries) < 2 {
		return nil, nil
	}
	p1, p2 := entries[0], entries[1]
	if p1.UserID == p2.UserID {
		return nil, ErrBadRequest
	}
	if ok, err := matchingPlayerReady(ctx, m.Repos, p1.UserID); err != nil {
		return nil, err
	} else if !ok {
		return nil, removeQueueEntries(ctx, m.Repos, []uuid.UUID{p1.ID})
	}
	if ok, err := matchingPlayerReady(ctx, m.Repos, p2.UserID); err != nil {
		return nil, err
	} else if !ok {
		return nil, removeQueueEntries(ctx, m.Repos, []uuid.UUID{p2.ID})
	}
	now := m.now()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusWaitingSecret,
		Player1ID: p1.UserID, Player2ID: p2.UserID,
		CurrentTurn: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := m.Games.Create(ctx, game); err != nil {
		return nil, err
	}
	if err := removeQueueEntries(ctx, m.Repos, []uuid.UUID{p1.ID, p2.ID}); err != nil {
		return nil, err
	}
	return game, nil
}

func matchingPlayerReady(ctx context.Context, repo repository.Repos, userID uuid.UUID) (bool, error) {
	user, err := repo.User.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil || user.IsDeleted() || user.IsMaster() {
		return false, nil
	}
	active, err := findActiveGameForUser(ctx, repo, userID)
	if err != nil {
		return false, err
	}
	return active == nil, nil
}

func removeQueueEntries(ctx context.Context, repos repository.Repos, ids []uuid.UUID) error {
	return repos.MatchingQueue.DeleteByIDs(ctx, ids)
}

func userWaitingInMatchingQueue(ctx context.Context, repo repository.Repos, userID uuid.UUID) (bool, error) {
	entry, err := repo.MatchingQueue.FindByUserID(ctx, userID)
	if err != nil || entry == nil {
		return false, err
	}
	return entry.Status == model.MatchingQueueWaiting, nil
}

func findActiveGameForUser(ctx context.Context, repo repository.Repos, userID uuid.UUID) (*model.Game, error) {
	games, err := repo.Game.ListByPlayerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		if game.Status == model.GameStatusWaitingSecret || game.Status == model.GameStatusInProgress {
			return game, nil
		}
	}
	return nil, nil
}

func NewMatchingUseCase(repos repository.Repos, notifier IEventNotifier) *MatchingUseCase {
	return &MatchingUseCase{
		Users:         repos.User,
		Games:         repos.Game,
		MatchingQueue: repos.MatchingQueue,
		Repos:         repos,
		Notifier:      notifier,
	}
}
