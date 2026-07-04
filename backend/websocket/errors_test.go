package websocket

import (
	"errors"
	"testing"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func TestWSErrorCodeAllBranches(t *testing.T) {
	cases := map[error]string{
		usecase.ErrUnauthorized:        "unauthorized",
		usecase.ErrTokenExpired:        "token_expired",
		usecase.ErrForbidden:           "forbidden",
		usecase.ErrNotFound:            "not_found",
		usecase.ErrDuplicateUser:       "conflict",
		usecase.ErrUserInActiveGame:    "conflict",
		usecase.ErrAlreadyInMatching:   "conflict",
		usecase.ErrGameNotStarted:      "conflict",
		usecase.ErrGameAlreadyFinished: "conflict",
		usecase.ErrNotYourTurn:         "conflict",
		usecase.ErrGameAlreadyStarted:  "conflict",
		usecase.ErrRateLimitExceeded:   "rate_limit_exceeded",
		usecase.ErrBadRequest:          "validation_error",
		model.ErrBadUsername:           "validation_error",
	}
	for err, want := range cases {
		code, _ := wsErrorCode(err)
		if code != want {
			t.Fatalf("%v code=%s want %s", err, code, want)
		}
	}
	code, msg := wsErrorCode(errors.New("boom"))
	if code != "internal_error" || msg != "internal server error" {
		t.Fatalf("default: %s %s", code, msg)
	}
}

func TestWSErrorCode(t *testing.T) {
	code, _ := wsErrorCode(usecase.ErrUnauthorized)
	if code != "unauthorized" {
		t.Fatalf("code=%s", code)
	}
	code, _ = wsErrorCode(usecase.ErrNotYourTurn)
	if code != "conflict" {
		t.Fatalf("code=%s", code)
	}
	code, _ = wsErrorCode(model.ErrDuplicateDigit)
	if code != "validation_error" {
		t.Fatalf("code=%s", code)
	}
}
