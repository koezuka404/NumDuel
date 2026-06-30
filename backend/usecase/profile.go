package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

//プロフィール・履歴取得ユースケース。
type IProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*GetProfileOutput, error)
	GetMatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]MatchHistoryItem, int64, error)
	GetLoginHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]LoginHistoryItem, int64, error)
	GetWSHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]WSConnectionHistoryItem, int64, error)
}

type ProfileUseCase struct {
	Users        repository.IUserRepo
	Rankings     repository.IRankingRepo
	MatchHistory repository.IMatchHistoryRepo
	LoginLogs    repository.ILoginLogRepo
	WSLogs       repository.IWSConnectionLogRepo
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

func (p *ProfileUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*GetProfileOutput, error) {
	user, err := p.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted() {
		return nil, ErrUnauthorized
	}
	out := &GetProfileOutput{Username: user.Username, WinCount: user.WinCount}
	rankings, err := p.Rankings.ListAll(ctx)
	if err != nil {
		return nil, err
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

func (p *ProfileUseCase) GetMatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]MatchHistoryItem, int64, error) {
	user, err := p.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, ErrUnauthorized
	}
	rows, total, err := p.MatchHistory.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
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

func (p *ProfileUseCase) GetLoginHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]LoginHistoryItem, int64, error) {
	user, err := p.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, ErrUnauthorized
	}
	rows, total, err := p.LoginLogs.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	items := make([]LoginHistoryItem, len(rows))
	for i, row := range rows {
		items[i] = LoginHistoryItem{Action: row.Action, CreatedAt: row.CreatedAt}
	}
	return items, total, nil
}

func (p *ProfileUseCase) GetWSHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]WSConnectionHistoryItem, int64, error) {
	user, err := p.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if user == nil || user.IsDeleted() {
		return nil, 0, ErrUnauthorized
	}
	rows, total, err := p.WSLogs.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	items := make([]WSConnectionHistoryItem, len(rows))
	for i, row := range rows {
		items[i] = WSConnectionHistoryItem{
			ConnectionID: row.ConnectionID, ConnectedAt: row.ConnectedAt, DisconnectedAt: row.DisconnectedAt,
		}
	}
	return items, total, nil
}

func NewProfileUseCase(repos repository.Repos) *ProfileUseCase {
	return &ProfileUseCase{
		Users:        repos.User,
		Rankings:     repos.Ranking,
		MatchHistory: repos.MatchHistory,
		LoginLogs:    repos.LoginLog,
		WSLogs:       repos.WSConnectionLog,
	}
}
