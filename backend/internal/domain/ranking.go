package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchHistory は勝敗履歴 Entity（読み取り専用モデル, 仕様書 4.4, 9.7）。
//
// 作成条件:
//   - guess_win でゲーム終了したときのみ FinishGameService が INSERT
//   - secret_setup_timeout では作成しない
//
// ユーザー名はスナップショットとして保存（後から username が変わっても履歴は不変）。
type MatchHistory struct {
	ID             uuid.UUID // PK
	GameID         uuid.UUID // UNIQUE（1 ゲーム 1 レコード）
	WinnerID       uuid.UUID
	LoserID        uuid.UUID
	WinnerUsername string // VARCHAR(50) スナップショット
	LoserUsername  string
	FinishedAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time // バックアップ差分同期用
}

// NewMatchHistory は FinishGameService 内で勝敗確定時に呼ぶファクトリ。
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

// Ranking はランキング Read Model（CQRS, 仕様書 4.5, 9.8 rankings テーブル）。
//
// 更新:
//   - 対戦 TX からは更新しない
//   - RebuildRankingUseCase / RankingRebuildWorker が users.win_count から全件再集計
//
// 除外: 削除済みユーザー、master ロール。
// UpdatedAt はバッチ完了時刻（UI「最終更新」表示用）。
type Ranking struct {
	UserID    uuid.UUID // PK（users.id と対応）
	Rank      int       // 1 始まり順位（win_count 降順）
	Username  string    // 非正規化（表示用）
	WinCount  int       // users.win_count のコピー
	UpdatedAt time.Time
}

// NewRanking は RebuildRankingUseCase が rankings 行を組み立てる際に使用。
func NewRanking(userID uuid.UUID, rank int, username string, winCount int, updatedAt time.Time) Ranking {
	return Ranking{
		UserID:    userID,
		Rank:      rank,
		Username:  username,
		WinCount:  winCount,
		UpdatedAt: updatedAt,
	}
}
