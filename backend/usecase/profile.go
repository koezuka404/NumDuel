package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

// ProfileDeps はプロフィール・履歴 UseCase の依存関係
type ProfileDeps struct {
	Repo repository.Repos
}

type GetProfileOutput struct {
	Username string
	WinCount int
	Rank     *int
}

type MatchHistoryItem struct {
	GameID         uuid.UUID
	WinnerUsername string
	LoserUsername  string
	FinishedAt     time.Time
}

type LoginHistoryItem struct {
	Action    model.LoginAction
	CreatedAt time.Time
}

type WSConnectionHistoryItem struct {
	ConnectionID   string
	ConnectedAt    time.Time
	DisconnectedAt *time.Time
}

func GetProfile(ctx context.Context, d ProfileDeps, userID uuid.UUID) (*GetProfileOutput, error) {
	user, err := d.Repo.User.FindByID(ctx, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, model.ErrUnauthorized()
	}
	out := &GetProfileOutput{Username: user.Username, WinCount: user.WinCount}
	rankings, err := d.Repo.Ranking.ListAll(ctx)
	if err != nil {
		return nil, model.ErrInternal("failed to load rankings")
	}
	for _, r := range rankings {
		if r.UserID == userID {
			rank := r.Rank
			out.Rank = &rank
			break
		}
	}
	return out, nil
}

func GetMatchHistory(ctx context.Context, d ProfileDeps, userID uuid.UUID, page, limit int) ([]MatchHistoryItem, int64, error) {
	user, err := d.Repo.User.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, model.ErrUnauthorized()
	}
	rows, total, err := d.Repo.MatchHistory.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to load match history")
	}
	items := make([]MatchHistoryItem, len(rows))
	for i, h := range rows {
		items[i] = MatchHistoryItem{
			GameID: h.GameID, WinnerUsername: h.WinnerUsername,
			LoserUsername: h.LoserUsername, FinishedAt: h.FinishedAt,
		}
	}
	return items, total, nil
}

func GetLoginHistory(ctx context.Context, d ProfileDeps, userID uuid.UUID, page, limit int) ([]LoginHistoryItem, int64, error) {
	user, err := d.Repo.User.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, model.ErrUnauthorized()
	}
	rows, total, err := d.Repo.LoginLog.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to load login history")
	}
	items := make([]LoginHistoryItem, len(rows))
	for i, l := range rows {
		items[i] = LoginHistoryItem{Action: l.Action, CreatedAt: l.CreatedAt}
	}
	return items, total, nil
}

func GetWSHistory(ctx context.Context, d ProfileDeps, userID uuid.UUID, page, limit int) ([]WSConnectionHistoryItem, int64, error) {
	user, err := d.Repo.User.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, model.ErrUnauthorized()
	}
	rows, total, err := d.Repo.WSConnectionLog.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to load ws history")
	}
	items := make([]WSConnectionHistoryItem, len(rows))
	for i, l := range rows {
		items[i] = WSConnectionHistoryItem{
			ConnectionID: l.ConnectionID, ConnectedAt: l.ConnectedAt, DisconnectedAt: l.DisconnectedAt,
		}
	}
	return items, total, nil
}
