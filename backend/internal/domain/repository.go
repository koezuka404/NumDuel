// Repository インターフェース群。UseCase は DB 操作をここ経由で行う。
package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Transaction は DB トランザクションのマーカー。
type Transaction interface{}

// TxManager はトランザクション境界を管理する。
type TxManager interface {
	Begin(ctx context.Context) (Transaction, error)
	Commit(tx Transaction) error
	Rollback(tx Transaction) error
}

// ActivityLog は activity_logs テーブルのレコード。
type ActivityLog struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	LogType   string
	Detail    json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoginLog は login_logs テーブルのレコード。
type LoginLog struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Action    LoginAction
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WSConnectionLog は ws_connection_logs テーブルのレコード。
type WSConnectionLog struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ConnectionID   string
	ConnectedAt    time.Time
	DisconnectedAt *time.Time
}

// RankingRebuildRow は RebuildRankingUseCase が users から集計する 1 行。
type RankingRebuildRow struct {
	UserID   uuid.UUID
	Username string
	WinCount int
}

// MatchingQueueEntry は matching_queue テーブルのレコード。
type MatchingQueueEntry struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Status    MatchingQueueStatus
	CreatedAt time.Time
}

// UserRepository は users テーブルへのアクセス。
type UserRepository interface {
	Create(ctx context.Context, tx Transaction, user *User) error
	Update(ctx context.Context, tx Transaction, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByEmailActive(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)
	CountMasters(ctx context.Context) (int64, error)
	ListForRankingRebuild(ctx context.Context) ([]RankingRebuildRow, error)
	IncrementWinCount(ctx context.Context, tx Transaction, userID uuid.UUID, now time.Time) error
	FindInactive(ctx context.Context, inactiveBefore time.Time) ([]*User, error)
	List(ctx context.Context, page, limit int) ([]*User, int64, error)
	Search(ctx context.Context, query string, page, limit int) ([]*User, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*User, error)
}

// GameRepository は games テーブルへのアクセス。
type GameRepository interface {
	Create(ctx context.Context, tx Transaction, game *Game) error
	Update(ctx context.Context, tx Transaction, game *Game) error
	FindByID(ctx context.Context, id uuid.UUID) (*Game, error)
	FindByIDForUpdate(ctx context.Context, tx Transaction, id uuid.UUID) (*Game, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*Game, error)
	FindAllInProgress(ctx context.Context) ([]*Game, error)
	FindWaitingSecretExpired(ctx context.Context, deadline time.Time) ([]*Game, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*Game, error)
}

// GuessRepository は guesses テーブルへのアクセス。
type GuessRepository interface {
	Create(ctx context.Context, tx Transaction, guess *Guess) error
	ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]Guess, error)
	CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]Guess, error)
}

// MatchHistoryRepository は match_histories テーブルへのアクセス。
type MatchHistoryRepository interface {
	Create(ctx context.Context, tx Transaction, history *MatchHistory) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]MatchHistory, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]MatchHistory, error)
}

// MatchingQueueRepository は matching_queue テーブルへのアクセス。
type MatchingQueueRepository interface {
	Insert(ctx context.Context, tx Transaction, entry *MatchingQueueEntry) error
	DeleteByIDs(ctx context.Context, tx Transaction, ids []uuid.UUID) error
	DeleteByUserID(ctx context.Context, tx Transaction, userID uuid.UUID) error
	FindWaitingForUpdate(ctx context.Context, tx Transaction, limit int) ([]MatchingQueueEntry, error)
	ExistsWaitingByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
	FindWaitingByUserID(ctx context.Context, userID uuid.UUID) (*MatchingQueueEntry, error)
}

// RankingRepository は rankings テーブルへのアクセス。
type RankingRepository interface {
	ReplaceAll(ctx context.Context, tx Transaction, rankings []Ranking) error
	ListAll(ctx context.Context) ([]Ranking, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]Ranking, error)
}

// RefreshTokenRepository は refresh_tokens テーブルへのアクセス。
type RefreshTokenRepository interface {
	Create(ctx context.Context, tx Transaction, token *RefreshToken) error
	Update(ctx context.Context, tx Transaction, token *RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeAllActiveByUserID(ctx context.Context, tx Transaction, userID uuid.UUID, now time.Time) error
	RevokeFamily(ctx context.Context, tx Transaction, familyID uuid.UUID, now time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// ActivityLogRepository は activity_logs テーブルへのアクセス。
type ActivityLogRepository interface {
	Create(ctx context.Context, log *ActivityLog) error
	Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]ActivityLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]ActivityLog, error)
}

// LoginLogRepository は login_logs テーブルへのアクセス。
type LoginLogRepository interface {
	Create(ctx context.Context, tx Transaction, log *LoginLog) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]LoginLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]LoginLog, error)
}

// WSConnectionLogRepository は ws_connection_logs テーブルへのアクセス。
type WSConnectionLogRepository interface {
	Create(ctx context.Context, log *WSConnectionLog) error
	UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]WSConnectionLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
}

// Repository は TxManager と各テーブル用 Repository へのアクセサを提供する。
type Repository interface {
	TxManager
	Users() UserRepository
	Games() GameRepository
	Guesses() GuessRepository
	MatchHistories() MatchHistoryRepository
	MatchingQueue() MatchingQueueRepository
	Rankings() RankingRepository
	RefreshTokens() RefreshTokenRepository
	ActivityLogs() ActivityLogRepository
	LoginLogs() LoginLogRepository
	WSConnectionLogs() WSConnectionLogRepository
}
