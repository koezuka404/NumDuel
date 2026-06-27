package domain

import (
	"time"

	"github.com/google/uuid"
)

// Game は対戦の集約ルート（Aggregate Root, games テーブル）。
//
// # 集約の責務
//
//   - 参加者・ターン・状態（Status）の整合性を Entity メソッドで保証
//   - 予想（Guess）の追加は AddGuess 経由のみ
//   - 秘密数字は HMAC ハッシュ文字列のみ保持（Player1Secret / Player2Secret）
//   - 平文 secret は Entity に載せない（SetSecretHash で UseCase が Hash 結果を渡す）
//
// # 先後手
//
//   - Player1ID … マッチング待機で先に登録したユーザー（先攻）
//   - Player2ID … 後から登録したユーザー（後攻）
//   - Start 後の current_turn_player_id は必ず player1 から開始
//
// # 状態と操作の対応
//
//   - WAITING_SECRET … SetSecretHash, CancelBySecretTimeout
//   - IN_PROGRESS    … AddGuess, Finish
//   - FINISHED       … 読み取りのみ
type Game struct {
	ID                  uuid.UUID  // PK
	Status              GameStatus // WAITING_SECRET / IN_PROGRESS / FINISHED
	Player1ID           uuid.UUID  // 先攻プレイヤー
	Player2ID           uuid.UUID  // 後攻プレイヤー
	Player1Secret       string     // player1 の HMAC ハッシュ（: 連結）。空 = 未登録
	Player2Secret       string     // player2 の HMAC ハッシュ。空 = 未登録
	CurrentTurnPlayerID *uuid.UUID // 現在ターンのプレイヤー。FINISHED 時は NULL
	CurrentTurn         int        // 1 始まりの通しターン番号
	WinnerID            *uuid.UUID // 勝者。secret_setup_timeout 時は NULL
	StartedAt           *time.Time // IN_PROGRESS 遷移時刻
	FinishedAt          *time.Time // FINISHED 遷移時刻
	CreatedAt           time.Time  // 秘密数字期限（SECRET_SETUP_SECONDS）の起点
	UpdatedAt           time.Time
}

// IsParticipant は userID が player1 / player2 のいずれかか。
func (g *Game) IsParticipant(userID uuid.UUID) bool {
	if g == nil {
		return false
	}
	return g.Player1ID == userID || g.Player2ID == userID
}

// IsCurrentTurn は userID が現在ターンのプレイヤーか。
func (g *Game) IsCurrentTurn(userID uuid.UUID) bool {
	if g == nil || g.CurrentTurnPlayerID == nil {
		return false
	}
	return *g.CurrentTurnPlayerID == userID
}

// CanGuess は GUESS / 自動予想を受け付け可能か。
//
// 条件: status == IN_PROGRESS かつ 自分のターン。
func (g *Game) CanGuess(userID uuid.UUID) bool {
	if g == nil {
		return false
	}
	return g.Status == GameStatusInProgress && g.IsCurrentTurn(userID)
}

// BothSecretsSet は両プレイヤーの秘密数字ハッシュが登録済みか。
// true になった時点で StartGameUseCase が Start を呼ぶ。
func (g *Game) BothSecretsSet() bool {
	if g == nil {
		return false
	}
	return g.Player1Secret != "" && g.Player2Secret != ""
}

// SetSecretHash はプレイヤーの秘密数字ハッシュを 1 回だけ登録（SetSecretNumberUseCase）。
//
// 前提:
//   - status == WAITING_SECRET
//   - 参加者であること
//   - 自分の secret が未登録
//
// hash は Infrastructure.SecretHashService が生成。平文は UseCase スコープ外へ持ち出さない。
func (g *Game) SetSecretHash(playerID uuid.UUID, hash string) error {
	if g == nil {
		return errForbidden("game is nil")
	}
	if g.Status != GameStatusWaitingSecret {
		return errGameNotStarted()
	}
	if !g.IsParticipant(playerID) {
		return errForbidden("not a participant")
	}
	if g.Player1ID == playerID {
		if g.Player1Secret != "" {
			return errValidation("secret already registered")
		}
		g.Player1Secret = hash
		return nil
	}
	if g.Player2Secret != "" {
		return errValidation("secret already registered")
	}
	g.Player2Secret = hash
	return nil
}

// SetSecret は SetSecretHash の別名。
func (g *Game) SetSecret(playerID uuid.UUID, hash string) error {
	return g.SetSecretHash(playerID, hash)
}

// OpponentID は対戦相手の user ID を返す。
func (g *Game) OpponentID(userID uuid.UUID) (uuid.UUID, error) {
	if g.Player1ID == userID {
		return g.Player2ID, nil
	}
	if g.Player2ID == userID {
		return g.Player1ID, nil
	}
	return uuid.Nil, errForbidden("not a participant")
}

// OpponentSecretHash は予想照合対象の相手ハッシュと playerSlot(1|2) を返す。
// SubmitGuessUseCase が SecretHashService.Verify に渡す。
func (g *Game) OpponentSecretHash(userID uuid.UUID) (hash string, opponentSlot int, err error) {
	if g.Player1ID == userID {
		return g.Player2Secret, 2, nil
	}
	if g.Player2ID == userID {
		return g.Player1Secret, 1, nil
	}
	return "", 0, errForbidden("not a participant")
}

// PlayerSlot は参加者のスロット番号（1=player1, 2=player2）。
// SecretHashService.Hash の playerSlot 引数に使用。
func (g *Game) PlayerSlot(userID uuid.UUID) (int, error) {
	if g.Player1ID == userID {
		return 1, nil
	}
	if g.Player2ID == userID {
		return 2, nil
	}
	return 0, errForbidden("not a participant")
}

// Start は両者の秘密数字登録完了後に IN_PROGRESS へ遷移（StartGameUseCase）。
//
// 副作用:
//   - status = IN_PROGRESS
//   - current_turn = 1
//   - current_turn_player_id = player1_id（先攻必須）
//   - started_at = now
//
// Redis ターン期限・WS 通知は COMMIT 後に UseCase が行う。
func (g *Game) Start(now time.Time) error {
	if g == nil {
		return errForbidden("game is nil")
	}
	if g.Status != GameStatusWaitingSecret || !g.BothSecretsSet() {
		return errValidation("cannot start game")
	}
	g.Status = GameStatusInProgress
	g.CurrentTurn = 1
	g.CurrentTurnPlayerID = &g.Player1ID
	g.StartedAt = &now
	g.UpdatedAt = now
	return nil
}

// advanceTurn は未勝利時のターン交代（内部用）。
// current_turn を +1 し、手番を相手に移す。
func (g *Game) advanceTurn(now time.Time) {
	if g.CurrentTurnPlayerID == nil {
		return
	}
	if *g.CurrentTurnPlayerID == g.Player1ID {
		g.CurrentTurnPlayerID = &g.Player2ID
	} else {
		g.CurrentTurnPlayerID = &g.Player1ID
	}
	g.CurrentTurn++
	g.UpdatedAt = now
}

// AddGuess は予想を追加し、未勝利ならターンを進める。
//
// 流れ:
//  1. CanGuess 検証（失敗時 game_not_started / game_already_finished / not_your_turn）
//  2. Guess Entity を生成（turn = current_turn のスナップショット）
//  3. IsWin(results) == false なら advanceTurn
//  4. IsWin == true ならターンは進めない（Finish は UseCase が FinishGameService で実行）
//
// Repository への INSERT は UseCase の TX 内で行う。
func (g *Game) AddGuess(
	playerID uuid.UUID,
	number GuessNumber,
	results [4]DigitResult,
	isAuto bool,
	now time.Time,
) (Guess, error) {
	if !g.CanGuess(playerID) {
		if g.Status == GameStatusWaitingSecret {
			return Guess{}, errGameNotStarted()
		}
		if g.Status == GameStatusFinished {
			return Guess{}, errGameAlreadyFinished()
		}
		return Guess{}, errNotYourTurn()
	}
	guess := NewGuess(g.ID, playerID, g.CurrentTurn, number, results, isAuto, now)
	if !IsWin(results) {
		g.advanceTurn(now)
	}
	return guess, nil
}

// Finish は guess_win による終了（FinishGameService）。
//
//   - status = FINISHED
//   - winner_id = winnerID
//   - current_turn_player_id = NULL
//   - finished_at = now
//
// match_histories 作成・win_count 加算は UseCase / FinishGameService が同一 TX で行う。
func (g *Game) Finish(winnerID uuid.UUID, now time.Time) error {
	if g == nil {
		return errForbidden("game is nil")
	}
	if g.Status == GameStatusFinished {
		return errGameAlreadyFinished()
	}
	if !g.IsParticipant(winnerID) {
		return errForbidden("winner is not a participant")
	}
	g.Status = GameStatusFinished
	g.WinnerID = &winnerID
	g.CurrentTurnPlayerID = nil
	g.FinishedAt = &now
	g.UpdatedAt = now
	return nil
}

// CancelBySecretTimeout は秘密数字登録期限切れによる終了。
//
//   - status = FINISHED
//   - winner_id = NULL（勝者なし）
//   - MatchHistory は作成しない
func (g *Game) CancelBySecretTimeout(now time.Time) error {
	if g == nil {
		return errForbidden("game is nil")
	}
	if g.Status != GameStatusWaitingSecret {
		return errValidation("game is not waiting for secrets")
	}
	g.Status = GameStatusFinished
	g.WinnerID = nil
	g.CurrentTurnPlayerID = nil
	g.FinishedAt = &now
	g.UpdatedAt = now
	return nil
}
