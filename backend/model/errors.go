package model

import "errors"

var (
	ErrBadUsername       = errors.New("username must be 3-50 alphanumeric/underscore characters")
	ErrBadEmail          = errors.New("email must be 1-255 characters")
	ErrWeakPassword      = errors.New("password must be at least 8 characters")
	ErrBadLoginEmail     = errors.New("email format is invalid")
	ErrBadRole           = errors.New("invalid role")
	ErrBadGameStatus     = errors.New("invalid game status")
	ErrBadDigit          = errors.New("digits must be numeric")
	ErrBadDigitLength    = errors.New("must be exactly 4 digits")
	ErrDuplicateDigit    = errors.New("digits must not repeat")
	ErrBadMatchingStatus = errors.New("invalid matching queue status")
	ErrBadLoginAction    = errors.New("invalid login action")
	ErrBadRefreshToken   = errors.New("invalid refresh token status")

	ErrUnauthorized        = errors.New("invalid credentials")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrGameNotStarted      = errors.New("game has not started")
	ErrGameAlreadyFinished = errors.New("game is already finished")
	ErrGameAlreadyStarted  = errors.New("game already started")
	ErrDuplicateUser       = errors.New("username or email already exists")
	ErrUserInActiveGame    = errors.New("user is already in an active game")
	ErrAlreadyInMatching   = errors.New("user is already in matching queue")
	ErrUserAlreadyDeleted  = errors.New("user is already deleted")
	ErrCannotDeleteSelf    = errors.New("cannot delete yourself")
	ErrCannotDeleteMaster  = errors.New("cannot delete master user")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrTokenExpired        = errors.New("access token expired")
)
