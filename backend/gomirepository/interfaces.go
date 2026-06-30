package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// IUserRepository は users テーブルへのアクセス
type IUserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ListAll(ctx context.Context) ([]*model.User, error)
	List(ctx context.Context, page, limit int) ([]*model.User, int64, error)
	Search(ctx context.Context, query string, page, limit int) ([]*model.User, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.User, error)
	ListInactiveSince(ctx context.Context, before time.Time) ([]*model.User, error) // AutoLogoutWorker 用
	TouchLastActivity(ctx context.Context, userID uuid.UUID, at time.Time) error // ActivityUpdateMiddleware 用（単一 UPDATE）
	ExistsActiveMaster(ctx context.Context) (bool, error)
}

// IGameRepository は games テーブルへのアクセス
type IGameRepository interface {
	Create(ctx context.Context, game *model.Game) error
	Update(ctx context.Context, game *model.Game) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error)
	FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListByPlayerID(ctx context.Context, userID uuid.UUID) ([]*model.Game, error)
	ListByStatus(ctx context.Context, status model.GameStatus) ([]*model.Game, error)
	ListByStatusCreatedBefore(ctx context.Context, status model.GameStatus, before time.Time) ([]*model.Game, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.Game, error)
}

// IGuessRepository は guesses テーブルへのアクセス
type IGuessRepository interface {
	Create(ctx context.Context, guess *model.Guess) error
	ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]model.Guess, error)
	CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Guess, error)
}

// IMatchHistoryRepository は match_histories テーブルへのアクセス
type IMatchHistoryRepository interface {
	Create(ctx context.Context, history *model.MatchHistory) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.MatchHistory, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.MatchHistory, error)
}

// IMatchingQueueRepository は matching_queue テーブルへのアクセス
type IMatchingQueueRepository interface {
	Insert(ctx context.Context, entry *model.MatchingQueueEntry) error
	DeleteByIDs(ctx context.Context, ids []uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	ListByStatusForUpdate(ctx context.Context, status model.MatchingQueueStatus, limit int) ([]model.MatchingQueueEntry, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.MatchingQueueEntry, error)
}

// IRankingRepository は rankings テーブルへのアクセス
type IRankingRepository interface {
	ReplaceAll(ctx context.Context, rankings []model.Ranking) error
	ListAll(ctx context.Context) ([]model.Ranking, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Ranking, error)
}

// IRefreshTokenRepository は refresh_tokens テーブルへのアクセス
type IRefreshTokenRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	FindByTokenHashWithUser(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	FindByTokenHashWithUserForUpdate(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error
	Create(ctx context.Context, token *model.RefreshToken) error
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// IActivityLogRepository は activity_logs テーブルへのアクセス
type IActivityLogRepository interface {
	Create(ctx context.Context, log *model.ActivityLog) error
	Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]model.ActivityLog, int64, error)
	ListDistinctLogTypes(ctx context.Context) ([]string, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.ActivityLog, error)
}

// ILoginLogRepository は login_logs テーブルへのアクセス
type ILoginLogRepository interface {
	Create(ctx context.Context, log *model.LoginLog) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.LoginLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.LoginLog, error)
}

// IWSConnectionLogRepository は ws_connection_logs テーブルへのアクセス
type IWSConnectionLogRepository interface {
	Create(ctx context.Context, log *model.WSConnectionLog) error
	UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.WSConnectionLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
}

// IRepository は各テーブル用 Repository へのアクセサを提供する
type IRepository interface {
	Users() IUserRepository
	Games() IGameRepository
	Guesses() IGuessRepository
	MatchHistories() IMatchHistoryRepository
	MatchingQueue() IMatchingQueueRepository
	Rankings() IRankingRepository
	RefreshTokens() IRefreshTokenRepository
	ActivityLogs() IActivityLogRepository
	LoginLogs() ILoginLogRepository
	WSConnectionLogs() IWSConnectionLogRepository
}
