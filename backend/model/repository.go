// Repository インターフェース群。UseCase は DB 操作をここ経由で行う。
package model

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
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID      `gorm:"type:uuid"`
	LogType   string          `gorm:"size:50;not null;index:idx_activity_logs_type_created,priority:1"`
	Detail    json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"not null;index:idx_activity_logs_type_created,priority:2"`
	UpdatedAt time.Time       `gorm:"not null"`
}

func (ActivityLog) TableName() string { return "activity_logs" }

// LoginLog は login_logs テーブルのレコード。
type LoginLog struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null"`
	Action    LoginAction `gorm:"size:20;not null"`
	CreatedAt time.Time   `gorm:"not null"`
	UpdatedAt time.Time   `gorm:"not null"`
}

func (LoginLog) TableName() string { return "login_logs" }

// WSConnectionLog は ws_connection_logs テーブルのレコード。
type WSConnectionLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	ConnectionID   string     `gorm:"size:64;not null"`
	ConnectedAt    time.Time  `gorm:"not null"`
	DisconnectedAt *time.Time
}

func (WSConnectionLog) TableName() string { return "ws_connection_logs" }

// RankingRebuildRow は RebuildRankingUseCase が users から集計する 1 行。
type RankingRebuildRow struct {
	UserID   uuid.UUID
	Username string
	WinCount int
}

// MatchingQueueEntry は matching_queue テーブルのレコード。
type MatchingQueueEntry struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID           `gorm:"type:uuid;not null"`
	Status    MatchingQueueStatus `gorm:"size:20;not null;index:idx_matching_queue_status_created,priority:1"`
	CreatedAt time.Time           `gorm:"not null;index:idx_matching_queue_status_created,priority:2"`
}

func (MatchingQueueEntry) TableName() string { return "matching_queue" }

// UserRepository は users テーブルへのアクセス。
type UserRepository interface {
	Create(ctx context.Context, tx Transaction, user *User) error
	Update(ctx context.Context, tx Transaction, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ListAll(ctx context.Context) ([]*User, error)
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
	ListByPlayerID(ctx context.Context, userID uuid.UUID) ([]*Game, error)
	ListByStatus(ctx context.Context, status GameStatus) ([]*Game, error)
	ListByStatusCreatedBefore(ctx context.Context, status GameStatus, before time.Time) ([]*Game, error)
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
	ListByStatusForUpdate(ctx context.Context, tx Transaction, status MatchingQueueStatus, limit int) ([]MatchingQueueEntry, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*MatchingQueueEntry, error)
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
	UpdateStatusByUserID(ctx context.Context, tx Transaction, userID uuid.UUID, fromStatus, toStatus RefreshTokenStatus, revokedAt *time.Time, now time.Time) error
	UpdateStatusByFamilyID(ctx context.Context, tx Transaction, familyID uuid.UUID, fromStatus, toStatus RefreshTokenStatus, revokedAt *time.Time, now time.Time) error
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
