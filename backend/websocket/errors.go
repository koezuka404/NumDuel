package websocket

import (
	"errors"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func wsErrorCode(err error) (code, message string) {
	switch {
	case errors.Is(err, usecase.ErrUnauthorized):
		return "unauthorized", err.Error()
	case errors.Is(err, usecase.ErrTokenExpired):
		return "token_expired", err.Error()
	case errors.Is(err, usecase.ErrForbidden):
		return "forbidden", err.Error()
	case errors.Is(err, usecase.ErrNotFound):
		return "not_found", err.Error()
	case errors.Is(err, usecase.ErrDuplicateUser),
		errors.Is(err, usecase.ErrUserInActiveGame),
		errors.Is(err, usecase.ErrAlreadyInMatching),
		errors.Is(err, usecase.ErrGameNotStarted),
		errors.Is(err, usecase.ErrGameAlreadyFinished),
		errors.Is(err, usecase.ErrNotYourTurn),
		errors.Is(err, usecase.ErrGameAlreadyStarted):
		return "conflict", err.Error()
	case errors.Is(err, usecase.ErrRateLimitExceeded):
		return "rate_limit_exceeded", err.Error()
	case errors.Is(err, usecase.ErrBadRequest),
		errors.Is(err, model.ErrBadUsername),
		errors.Is(err, model.ErrBadEmail),
		errors.Is(err, model.ErrBadLoginEmail),
		errors.Is(err, model.ErrWeakPassword),
		errors.Is(err, model.ErrBadDigit),
		errors.Is(err, model.ErrBadDigitLength),
		errors.Is(err, model.ErrDuplicateDigit):
		return "validation_error", err.Error()
	default:
		return "internal_error", model.ErrInternal.Error()
	}
}
