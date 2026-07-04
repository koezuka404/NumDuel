package usecase

import (
	"errors"
	"testing"

	"github.com/numduel/numduel/model"
)

func TestValidateUsername(t *testing.T) {
	if err := ValidateUsername("alice"); err != nil {
		t.Fatalf("valid username: %v", err)
	}
	if err := ValidateUsername("ab"); !errors.Is(err, model.ErrBadUsername) {
		t.Fatalf("short username: %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("password123"); err != nil {
		t.Fatalf("valid password: %v", err)
	}
	if err := ValidatePassword("short"); !errors.Is(err, model.ErrWeakPassword) {
		t.Fatalf("weak password: %v", err)
	}
}

func TestValidateLoginEmail(t *testing.T) {
	if err := ValidateLoginEmail("user@test.local"); err != nil {
		t.Fatalf("valid email: %v", err)
	}
	if err := ValidateLoginEmail("invalid"); !errors.Is(err, model.ErrBadLoginEmail) {
		t.Fatalf("invalid email: %v", err)
	}
	if err := ValidateLoginEmail(""); !errors.Is(err, model.ErrBadLoginEmail) {
		t.Fatalf("empty email: %v", err)
	}
	if err := ValidateLoginEmail("@test.local"); !errors.Is(err, model.ErrBadLoginEmail) {
		t.Fatalf("@ prefix: %v", err)
	}
	if err := ValidateLoginEmail("user@"); !errors.Is(err, model.ErrBadLoginEmail) {
		t.Fatalf("@ suffix: %v", err)
	}
	if err := ValidateLoginEmail(string(make([]byte, 51))); !errors.Is(err, model.ErrBadLoginEmail) {
		t.Fatalf("long email: %v", err)
	}
}

func TestValidateEmailEdgeCases(t *testing.T) {
	if err := ValidateEmail(""); !errors.Is(err, model.ErrBadEmail) {
		t.Fatalf("empty email: %v", err)
	}
	if err := ValidateEmail(string(make([]byte, 256))); !errors.Is(err, model.ErrBadEmail) {
		t.Fatalf("long email: %v", err)
	}
}

func TestValidateUsernameEdgeCases(t *testing.T) {
	if err := ValidateUsername("a!"); !errors.Is(err, model.ErrBadUsername) {
		t.Fatalf("invalid chars: %v", err)
	}
	if err := ValidateUsername(string(make([]byte, 51))); !errors.Is(err, model.ErrBadUsername) {
		t.Fatalf("long username: %v", err)
	}
}
