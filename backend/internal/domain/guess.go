package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecretNumber は秘密の 4 桁数字（値オブジェクト）。
//
// ルール:
//   - 4 桁固定
//   - 各桁 0〜9
//   - 同じ数字は不可（重複なし）
//
// 平文は WS で 1 回だけ送られ、UseCase 内で Hash 後に破棄する。
// Entity / DB に平文を保持しない。
type SecretNumber struct {
	digits [4]int // 非公開。Digits() / String() 経由でのみ参照
}

// NewSecretNumber は [4]int から SecretNumber を生成。
func NewSecretNumber(digits [4]int) (SecretNumber, error) {
	if err := parseFourDigitsArray(digits); err != nil {
		return SecretNumber{}, err
	}
	return SecretNumber{digits: digits}, nil
}

// NewSecretNumberFromString は WS SET_SECRET の平文文字列（"1234" 形式）から生成。
func NewSecretNumberFromString(s string) (SecretNumber, error) {
	digits, err := parseFourDigits(s)
	if err != nil {
		return SecretNumber{}, err
	}
	return SecretNumber{digits: digits}, nil
}

// Digits は内部 4 桁配列を返す。SecretHashService.Hash の入力に UseCase が使用。
func (s SecretNumber) Digits() [4]int {
	return s.digits
}

// String は 4 桁文字列化。ログ出力禁止。
func (s SecretNumber) String() string {
	return digitsToString(s.digits)
}

// GuessNumber は予想 4 桁数字（値オブジェクト）。
// SecretNumber と同一の桁ルール。SubmitGuess / HandleTimeout の入力検証に使う。
type GuessNumber struct {
	digits [4]int
}

// NewGuessNumber は [4]int から GuessNumber を生成。
func NewGuessNumber(digits [4]int) (GuessNumber, error) {
	if err := parseFourDigitsArray(digits); err != nil {
		return GuessNumber{}, err
	}
	return GuessNumber{digits: digits}, nil
}

// NewGuessNumberFromString は WS GUESS / 自動予想の文字列から生成。
func NewGuessNumberFromString(s string) (GuessNumber, error) {
	digits, err := parseFourDigits(s)
	if err != nil {
		return GuessNumber{}, err
	}
	return GuessNumber{digits: digits}, nil
}

func (g GuessNumber) Digits() [4]int {
	return g.digits
}

func (g GuessNumber) String() string {
	return digitsToString(g.digits)
}

// Guess は予想履歴 Entity（guesses テーブル）。
//
// 追加は必ず Game.AddGuess 経由。
// guess_number は自分の予想のみ平文保存。相手の予想内容は WS に含めない。
type Guess struct {
	ID           uuid.UUID     // PK
	GameID       uuid.UUID     // FK → games
	PlayerID     uuid.UUID     // FK → users（予想したプレイヤー）
	Turn         int           // games.current_turn のスナップショット
	GuessNumber  string        // 4 桁平文（VARCHAR(4)）
	DigitResults []DigitResult // JSONB [0,1,0,0] 形式
	HitCount     int           // digit_results 内の 1 の個数（0〜4）
	IsAuto       bool          // HandleTimeoutUseCase による自動予想なら true
	CreatedAt    time.Time
	UpdatedAt    time.Time     // バックアップ差分 UPSERT 用
}

// NewGuess は現在ターンの予想レコードを組み立てる。
//
// digitResults は Infrastructure.SecretHashService.Verify または Domain.JudgeDigits の結果。
// hitCount は IsWin 判定にも使用（4 なら勝利 → FinishGameService）。
func NewGuess(
	gameID, playerID uuid.UUID,
	turn int,
	number GuessNumber,
	results [4]DigitResult,
	isAuto bool,
	now time.Time,
) Guess {
	digitResults := make([]DigitResult, 4)
	copy(digitResults, results[:])
	hitCount := HitCount(results)
	return Guess{
		ID:           uuid.New(),
		GameID:       gameID,
		PlayerID:     playerID,
		Turn:         turn,
		GuessNumber:  number.String(),
		DigitResults: digitResults,
		HitCount:     hitCount,
		IsAuto:       isAuto,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// HitCount は判定結果配列内の当たり（1）の個数を数える。
func HitCount(results [4]DigitResult) int {
	n := 0
	for _, r := range results {
		if r == DigitHit {
			n++
		}
	}
	return n
}

// DigitResultsToInts は API / WS ペイロード用に []int{0,1,0,0} へ変換する。
func DigitResultsToInts(results []DigitResult) []int {
	out := make([]int, len(results))
	for i, r := range results {
		out[i] = int(r)
	}
	return out
}
