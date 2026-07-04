package usecase

import (
	"errors"
	"testing"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func TestMapValidationErr(t *testing.T) {
	cases := []struct {
		in   error
		want error
	}{
		{model.ErrBadUsername, ErrBadRequest},
		{model.ErrBadEmail, ErrBadRequest},
		{model.ErrBadLoginEmail, ErrBadRequest},
		{model.ErrWeakPassword, ErrBadRequest},
		{model.ErrBadDigit, ErrBadRequest},
		{model.ErrBadDigitLength, ErrBadRequest},
		{model.ErrDuplicateDigit, ErrBadRequest},
		{errors.New("other"), errors.New("other")},
	}
	for _, tc := range cases {
		got := mapValidationErr(tc.in)
		if !errors.Is(got, tc.want) && got.Error() != tc.want.Error() {
			t.Fatalf("mapValidationErr(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestMapRepoNotFound(t *testing.T) {
	got := mapRepoNotFound(repository.ErrNotFound)
	if !errors.Is(got, ErrNotFound) {
		t.Fatalf("mapRepoNotFound: %v", got)
	}
	other := errors.New("db error")
	if mapRepoNotFound(other) != other {
		t.Fatalf("mapRepoNotFound should pass through")
	}
}

func TestMapRepoNotFoundAs(t *testing.T) {
	target := ErrForbidden
	got := mapRepoNotFoundAs(repository.ErrNotFound, target)
	if !errors.Is(got, target) {
		t.Fatalf("mapRepoNotFoundAs: %v", got)
	}
}

func TestMapRepoNotFoundAsPassThrough(t *testing.T) {
	other := errors.New("db error")
	if mapRepoNotFoundAs(other, ErrForbidden) != other {
		t.Fatalf("mapRepoNotFoundAs should pass through")
	}
}

func TestIsRepoNotFound(t *testing.T) {
	if !isRepoNotFound(repository.ErrNotFound) {
		t.Fatalf("expected true")
	}
	if isRepoNotFound(errors.New("other")) {
		t.Fatalf("expected false")
	}
}

func TestIsTimeoutRaceError(t *testing.T) {
	for _, err := range []error{ErrNotYourTurn, ErrGameAlreadyFinished, ErrGameNotStarted, ErrNotFound} {
		if !isTimeoutRaceError(err) {
			t.Fatalf("expected race error for %v", err)
		}
	}
	if isTimeoutRaceError(ErrBadRequest) {
		t.Fatalf("unexpected race error")
	}
}
