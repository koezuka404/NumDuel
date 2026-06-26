package domain

import "fmt"

// ドメインエラーコード（仕様書 第8章）。
// Domain 層では HTTP ステータスを持たない。UseCase / Controller が code に応じて HTTP を決定する。
const (
	CodeValidation          = "validation_error"
	CodeInvalidDigitLength  = "invalid_digit_length"  // 4 桁でない
	CodeInvalidDigit        = "invalid_digit"         // 数字以外を含む
	CodeDuplicateDigit      = "duplicate_digit"       // 同じ数字の重複
	CodeGameNotStarted      = "game_not_started"      // WAITING_SECRET 中の GUESS 等
	CodeGameAlreadyFinished = "game_already_finished" // FINISHED 後の操作
	CodeNotYourTurn         = "not_your_turn"         // 自分のターンでない
	CodeForbidden           = "forbidden"             // 参加者でない等
)

// DomainError は Domain 層の業務エラー。
// Code は API レスポンス error.code と一致させる（仕様書 6.1.1）。
type DomainError struct {
	Code string
	Msg  string
}

func (e *DomainError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Code
}

// IsDomainError は err が *DomainError か判定する。テスト・UseCase のエラー変換用。
func IsDomainError(err error) (*DomainError, bool) {
	if err == nil {
		return nil, false
	}
	de, ok := err.(*DomainError)
	return de, ok
}

func newDomainError(code, msg string) *DomainError {
	return &DomainError{Code: code, Msg: msg}
}

func errValidation(msg string) *DomainError {
	return newDomainError(CodeValidation, msg)
}

func errInvalidDigitLength() *DomainError {
	return newDomainError(CodeInvalidDigitLength, "must be exactly 4 digits")
}

func errInvalidDigit() *DomainError {
	return newDomainError(CodeInvalidDigit, "digits must be numeric")
}

func errDuplicateDigit() *DomainError {
	return newDomainError(CodeDuplicateDigit, "digits must not repeat")
}

func errGameNotStarted() *DomainError {
	return newDomainError(CodeGameNotStarted, "game has not started")
}

func errGameAlreadyFinished() *DomainError {
	return newDomainError(CodeGameAlreadyFinished, "game is already finished")
}

func errNotYourTurn() *DomainError {
	return newDomainError(CodeNotYourTurn, "not your turn")
}

func errForbidden(msg string) *DomainError {
	if msg == "" {
		msg = "forbidden"
	}
	return newDomainError(CodeForbidden, msg)
}

// parseFourDigits は文字列から 4 桁数字（0-9, 重複なし）を検証する。
// SecretNumber / GuessNumber の共通バリデーション（仕様書 1.1, 4.1）。
func parseFourDigits(input string) ([4]int, error) {
	if len(input) != 4 {
		return [4]int{}, errInvalidDigitLength()
	}
	seen := map[int]struct{}{}
	var digits [4]int
	for i, ch := range input {
		if ch < '0' || ch > '9' {
			return [4]int{}, errInvalidDigit()
		}
		d := int(ch - '0')
		if _, ok := seen[d]; ok {
			return [4]int{}, errDuplicateDigit()
		}
		seen[d] = struct{}{}
		digits[i] = d
	}
	return digits, nil
}

// parseFourDigitsArray は [4]int 配列版の検証。NewSecretNumber / NewGuessNumber 用。
func parseFourDigitsArray(digits [4]int) error {
	seen := map[int]struct{}{}
	for i := 0; i < 4; i++ {
		if digits[i] < 0 || digits[i] > 9 {
			return errInvalidDigit()
		}
		if _, ok := seen[digits[i]]; ok {
			return errDuplicateDigit()
		}
		seen[digits[i]] = struct{}{}
	}
	return nil
}

func digitsToString(digits [4]int) string {
	return fmt.Sprintf("%d%d%d%d", digits[0], digits[1], digits[2], digits[3])
}
