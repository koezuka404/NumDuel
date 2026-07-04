//API共通のJSONレスポンス形式を提供する
package dto

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func WriteData(c echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

func WriteError(c echo.Context, err error) error {
	status, code, msg := mapHTTPError(err)
	return c.JSON(status, map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}

func mapHTTPError(err error) (int, string, string) {
	switch {
	case errors.Is(err, usecase.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized", err.Error()
	case errors.Is(err, usecase.ErrTokenExpired):
		return http.StatusNotFound, "token_expired", err.Error()
	case errors.Is(err, usecase.ErrForbidden):
		return http.StatusForbidden, "forbidden", err.Error()
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound, "not_found", err.Error()
	case errors.Is(err, usecase.ErrDuplicateUser),
		errors.Is(err, usecase.ErrUserInActiveGame),
		errors.Is(err, usecase.ErrAlreadyInMatching),
		errors.Is(err, usecase.ErrGameNotStarted),
		errors.Is(err, usecase.ErrGameAlreadyFinished),
		errors.Is(err, usecase.ErrNotYourTurn),
		errors.Is(err, usecase.ErrGameAlreadyStarted),
		errors.Is(err, usecase.ErrUserAlreadyDeleted):
		return http.StatusConflict, conflictCode(err), err.Error()
	case errors.Is(err, usecase.ErrCannotDeleteSelf),
		errors.Is(err, usecase.ErrCannotDeleteMaster):
		return http.StatusForbidden, conflictCode(err), err.Error()
	case errors.Is(err, usecase.ErrRateLimitExceeded):
		return http.StatusTooManyRequests, "rate_limit_exceeded", err.Error()
	case errors.Is(err, usecase.ErrBadRequest),
		errors.Is(err, model.ErrBadUsername),
		errors.Is(err, model.ErrBadEmail),
		errors.Is(err, model.ErrBadLoginEmail),
		errors.Is(err, model.ErrWeakPassword),
		errors.Is(err, model.ErrBadDigit),
		errors.Is(err, model.ErrBadDigitLength),
		errors.Is(err, model.ErrDuplicateDigit):
		return http.StatusBadRequest, "validation_error", err.Error()
	default:
		return http.StatusInternalServerError, "internal_error", model.ErrInternal.Error()
	}
}

func conflictCode(err error) string {
	switch {
	case errors.Is(err, usecase.ErrDuplicateUser):
		return "duplicate_user"
	case errors.Is(err, usecase.ErrUserInActiveGame):
		return "user_in_active_game"
	case errors.Is(err, usecase.ErrAlreadyInMatching):
		return "already_in_matching"
	case errors.Is(err, usecase.ErrGameNotStarted):
		return "game_not_started"
	case errors.Is(err, usecase.ErrGameAlreadyFinished):
		return "game_already_finished"
	case errors.Is(err, usecase.ErrNotYourTurn):
		return "not_your_turn"
	case errors.Is(err, usecase.ErrGameAlreadyStarted):
		return "game_already_started"
	case errors.Is(err, usecase.ErrUserAlreadyDeleted):
		return "user_already_deleted"
	case errors.Is(err, usecase.ErrCannotDeleteSelf):
		return "cannot_delete_self"
	case errors.Is(err, usecase.ErrCannotDeleteMaster):
		return "cannot_delete_master"
	default:
		return "conflict"
	}
}
