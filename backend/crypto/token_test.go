package crypto

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func TestJWTServiceIssueAndParse(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	userID := uuid.New()
	now := time.Now().UTC()
	token, err := svc.Issue(userID, model.RoleUser, now)
	if err != nil || token == "" {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := svc.Parse(token)
	if err != nil || claims.UserID != userID || claims.Role != model.RoleUser {
		t.Fatalf("Parse: %+v err=%v", claims, err)
	}
}

func TestJWTServiceExpiredToken(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 1)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	token, err := svc.Issue(uuid.New(), model.RoleUser, time.Now().UTC().Add(-5*time.Minute))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	_, err = svc.Parse(token)
	if !errors.Is(err, usecase.ErrTokenExpired) {
		t.Fatalf("expected token_expired, got %v", err)
	}
}

func TestJWTServiceRejectsWrongSecret(t *testing.T) {
	svc1, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt1: %v", err)
	}
	svc2, err := NewJWTService("different-secret-key-at-least-32-chars!!", 60)
	if err != nil {
		t.Fatalf("jwt2: %v", err)
	}
	token, err := svc1.Issue(uuid.New(), model.RoleUser, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	_, err = svc2.Parse(token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("wrong secret: %v", err)
	}
}

func TestJWTServiceRejectsTamperedToken(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	_, err = svc.Parse("invalid.jwt.token")
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("invalid token: %v", err)
	}
}

func TestJWTServiceRejectsShortSecret(t *testing.T) {
	_, err := NewJWTService("short", 60)
	if err == nil {
		t.Fatalf("expected error for short secret")
	}
}
