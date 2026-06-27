package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type userModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Username       string     `gorm:"size:50;not null;uniqueIndex"`
	Email          string     `gorm:"size:255;not null;uniqueIndex"`
	Password       string     `gorm:"size:255;not null;column:password"`
	Role           string     `gorm:"size:20;not null"`
	WinCount       int        `gorm:"not null;default:0"`
	DeletedAt      *time.Time `gorm:"index"`
	DeletedBy      *uuid.UUID `gorm:"type:uuid"`
	LastActivityAt time.Time  `gorm:"not null"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (userModel) TableName() string { return "users" }

type gameModel struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Status              string     `gorm:"size:20;not null;index"`
	Player1ID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Player2ID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Player1Secret       *string    `gorm:"size:255"`
	Player2Secret       *string    `gorm:"size:255"`
	CurrentTurnPlayerID *uuid.UUID `gorm:"type:uuid"`
	CurrentTurn         int        `gorm:"not null;default:1"`
	WinnerID            *uuid.UUID `gorm:"type:uuid"`
	StartedAt           *time.Time
	FinishedAt          *time.Time
	CreatedAt           time.Time `gorm:"not null"`
	UpdatedAt           time.Time `gorm:"not null"`
}

func (gameModel) TableName() string { return "games" }

type guessModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	GameID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_guesses_game_turn_player;index:idx_guesses_game_turn,priority:1"`
	PlayerID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_guesses_game_turn_player"`
	Turn         int            `gorm:"not null;uniqueIndex:uq_guesses_game_turn_player;index:idx_guesses_game_turn,priority:2"`
	GuessNumber  string         `gorm:"size:4;not null"`
	DigitResults datatypes.JSON `gorm:"type:jsonb;not null"`
	HitCount     int            `gorm:"not null"`
	IsAuto       bool           `gorm:"not null;default:false"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
}

func (guessModel) TableName() string { return "guesses" }

type matchHistoryModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	GameID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	WinnerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	LoserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	WinnerUsername string    `gorm:"size:50;not null"`
	LoserUsername  string    `gorm:"size:50;not null"`
	FinishedAt     time.Time `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (matchHistoryModel) TableName() string { return "match_histories" }

type rankingModel struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Rank      int       `gorm:"not null;index"`
	Username  string    `gorm:"size:50;not null"`
	WinCount  int       `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (rankingModel) TableName() string { return "rankings" }

type matchingQueueModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Status    string    `gorm:"size:20;not null;index:idx_matching_queue_status_created,priority:1"`
	CreatedAt time.Time `gorm:"not null;index:idx_matching_queue_status_created,priority:2"`
}

func (matchingQueueModel) TableName() string { return "matching_queue" }

type activityLogModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID     `gorm:"type:uuid"`
	LogType   string         `gorm:"size:50;not null;index:idx_activity_logs_type_created,priority:1"`
	Detail    datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt time.Time      `gorm:"not null;index:idx_activity_logs_type_created,priority:2"`
	UpdatedAt time.Time      `gorm:"not null"`
}

func (activityLogModel) TableName() string { return "activity_logs" }

type loginLogModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Action    string    `gorm:"size:20;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (loginLogModel) TableName() string { return "login_logs" }

type wsConnectionLogModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	ConnectionID   string     `gorm:"size:64;not null"`
	ConnectedAt    time.Time  `gorm:"not null"`
	DisconnectedAt *time.Time
}

func (wsConnectionLogModel) TableName() string { return "ws_connection_logs" }

type refreshTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash string     `gorm:"size:255;not null;uniqueIndex"`
	FamilyID  uuid.UUID  `gorm:"type:uuid;not null"`
	Status    string     `gorm:"size:20;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }
