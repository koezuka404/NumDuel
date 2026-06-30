package model

import (
	"time"

	"github.com/google/uuid"
)

// MatchHistory は勝敗履歴 Entity（読み取り専用モデル）

// 作成条件:
// - guess_win でゲーム終了したときのみ FinishGameService が INSERT
// - secret_setup_timeout では作成しない

// ユーザー名はスナップショットとして保存（後から username が変わっても履歴は不変）
type MatchHistory struct {
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

func (MatchHistory) TableName() string { return "match_histories" }

// NewMatchHistory は FinishGameService 内で勝敗確定時に呼ぶファクトリ
func NewMatchHistory(
	gameID, winnerID, loserID uuid.UUID,
	winnerUsername, loserUsername string,
	finishedAt time.Time,
) MatchHistory {
	now := finishedAt
	return MatchHistory{
		ID:             uuid.New(),
		GameID:         gameID,
		WinnerID:       winnerID,
		LoserID:        loserID,
		WinnerUsername: winnerUsername,
		LoserUsername:  loserUsername,
		FinishedAt:     finishedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Ranking はランキング Read Model（CQRS, rankings テーブル）

// 更新:
// - 対戦 TX からは更新しない
// - RebuildRankingUseCase / RankingRebuildWorker が users.win_count から全件再集計

// 除外: 削除済みユーザー、master ロール
// UpdatedAt はバッチ完了時刻（UI「最終更新」表示用）
type Ranking struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Rank      int       `gorm:"not null;index"`
	Username  string    `gorm:"size:50;not null"`
	WinCount  int       `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Ranking) TableName() string { return "rankings" }

// NewRanking は RebuildRankingUseCase が rankings 行を組み立てる際に使用
func NewRanking(userID uuid.UUID, rank int, username string, winCount int, updatedAt time.Time) Ranking {
	return Ranking{
		UserID:    userID,
		Rank:      rank,
		Username:  username,
		WinCount:  winCount,
		UpdatedAt: updatedAt,
	}
}

// RankingRebuildRow は RebuildRankingUseCase が users から集計する 1 行
type RankingRebuildRow struct {
	UserID   uuid.UUID
	Username string
	WinCount int
}
