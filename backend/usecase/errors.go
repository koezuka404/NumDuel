package usecase

import (
	"errors"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

var (
	ErrUnauthorized       = model.ErrUnauthorized
	ErrForbidden          = model.ErrForbidden
	ErrNotFound           = model.ErrNotFound
	ErrDuplicateUser      = model.ErrDuplicateUser
	ErrBadRequest         = errors.New("bad request")
	ErrTokenExpired       = model.ErrTokenExpired
	ErrRateLimitExceeded  = model.ErrRateLimitExceeded
	ErrUserInActiveGame   = model.ErrUserInActiveGame
	ErrAlreadyInMatching  = model.ErrAlreadyInMatching
	ErrGameNotStarted     = model.ErrGameNotStarted
	ErrGameAlreadyStarted = model.ErrGameAlreadyStarted
	ErrGameAlreadyFinished = model.ErrGameAlreadyFinished
	ErrNotYourTurn        = model.ErrNotYourTurn
	ErrUserAlreadyDeleted = model.ErrUserAlreadyDeleted
	ErrCannotDeleteSelf   = model.ErrCannotDeleteSelf
	ErrCannotDeleteMaster = model.ErrCannotDeleteMaster
)

func mapRepoNotFound(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func mapRepoNotFoundAs(err error, target error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return target
	}
	return err
}

func isRepoNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

func mapValidationErr(err error) error {
	switch {
	case errors.Is(err, model.ErrBadUsername),
		errors.Is(err, model.ErrBadEmail),
		errors.Is(err, model.ErrBadLoginEmail),
		errors.Is(err, model.ErrWeakPassword),
		errors.Is(err, model.ErrBadDigit),
		errors.Is(err, model.ErrBadDigitLength),
		errors.Is(err, model.ErrDuplicateDigit):
		return ErrBadRequest
	default:
		return err
	}
}
