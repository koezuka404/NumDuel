// ドメイン層の業務エラーHTTP ステータスへの変換は dto が担当
package model

import "fmt"

// API レスポンス error.code と一致させる定数群
const (
	CodeValidation          = "validation_error"
	CodeInvalidDigitLength  = "invalid_digit_length"
	CodeInvalidDigit        = "invalid_digit"
	CodeDuplicateDigit      = "duplicate_digit"
	CodeGameNotStarted      = "game_not_started"
	CodeGameAlreadyFinished = "game_already_finished"
	CodeNotYourTurn         = "not_your_turn"
	CodeForbidden           = "forbidden"
	CodeDuplicateUser       = "duplicate_user"
	CodeUserInActiveGame    = "user_in_active_game"
	CodeAlreadyInMatching   = "already_in_matching"
	CodeUnauthorized        = "unauthorized"
	CodeTokenExpired        = "token_expired"
	CodeNotFound            = "not_found"
	CodeRateLimitExceeded   = "rate_limit_exceeded"
	CodeGameAlreadyStarted  = "game_already_started"
	CodeUserAlreadyDeleted  = "user_already_deleted"
	CodeCannotDeleteSelf    = "cannot_delete_self"
	CodeCannotDeleteMaster  = "cannot_delete_master"
	CodeInternalError       = "internal_error"
)

type DomainError struct {
	Code string
	Msg  string
}

// Error は message があればそれを、なければ code を返す
func (e *DomainError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Code
}

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

func ErrForbidden(msg string) *DomainError {
	return errForbidden(msg)
}

func ErrDuplicateUser() *DomainError {
	return newDomainError(CodeDuplicateUser, "username or email already exists")
}

func ErrUserInActiveGame() *DomainError {
	return newDomainError(CodeUserInActiveGame, "user is already in an active game")
}

func ErrAlreadyInMatching() *DomainError {
	return newDomainError(CodeAlreadyInMatching, "user is already in matching queue")
}

func ErrValidation(msg string) *DomainError {
	return errValidation(msg)
}

func ErrUnauthorized() *DomainError {
	return newDomainError(CodeUnauthorized, "invalid credentials")
}

func ErrTokenExpired() *DomainError {
	return newDomainError(CodeTokenExpired, "access token expired")
}

func ErrInternal(msg string) *DomainError {
	if msg == "" {
		msg = "internal server error"
	}
	return newDomainError(CodeInternalError, msg)
}

func ErrNotFound(msg string) *DomainError {
	if msg == "" {
		msg = "not found"
	}
	return newDomainError(CodeNotFound, msg)
}

func ErrRateLimitExceeded() *DomainError {
	return newDomainError(CodeRateLimitExceeded, "rate limit exceeded")
}

func ErrGameAlreadyStarted() *DomainError {
	return newDomainError(CodeGameAlreadyStarted, "game already started")
}

func ErrGameAlreadyFinished() *DomainError {
	return errGameAlreadyFinished()
}

func ErrUserAlreadyDeleted() *DomainError {
	return newDomainError(CodeUserAlreadyDeleted, "user is already deleted")
}

func ErrCannotDeleteSelf() *DomainError {
	return newDomainError(CodeCannotDeleteSelf, "cannot delete yourself")
}

func ErrCannotDeleteMaster() *DomainError {
	return newDomainError(CodeCannotDeleteMaster, "cannot delete master user")
}

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
