package crypto

import (
	"errors"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func TestNewJWTServiceInvalidExpiry(t *testing.T) {
	_, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 0)
	if err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestJWTParseInvalidSubject(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	claims := &accessClaims{
		Role: string(model.RoleUser),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "not-a-uuid",
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(svc.secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = svc.Parse(token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("parse: %v", err)
	}
}

func TestJWTParseEmptyJTI(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	claims := &accessClaims{
		Role: string(model.RoleUser),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.New().String(),
			ID:        "",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(svc.secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = svc.Parse(token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("parse: %v", err)
	}
}

func TestRefreshTokenGenerate(t *testing.T) {
	svc := NewRefreshTokenService()
	pair, err := svc.Generate()
	if err != nil || pair.Plaintext == "" || pair.Hash == "" {
		t.Fatalf("generate: %+v err=%v", pair, err)
	}
	if svc.Hash(pair.Plaintext) != pair.Hash {
		t.Fatal("hash mismatch")
	}
}

func TestGenerateGuessNumber(t *testing.T) {
	svc := NewRandomNumberService()
	n, err := svc.GenerateGuessNumber()
	if err != nil || len(n) != 4 {
		t.Fatalf("guess: %q err=%v", n, err)
	}
}

func TestPasswordHashError(t *testing.T) {
	orig := generateFromPasswordFn
	t.Cleanup(func() { generateFromPasswordFn = orig })
	generateFromPasswordFn = func([]byte, int) ([]byte, error) {
		return nil, errors.New("bcrypt failed")
	}
	_, err := NewPasswordService().Hash("password123")
	if err == nil {
		t.Fatal("expected hash error")
	}
}

func TestRandIntCryptoError(t *testing.T) {
	orig := intRandFn
	t.Cleanup(func() { intRandFn = orig })
	intRandFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, errors.New("rand failed")
	}
	_, err := randInt(0, 9)
	if err == nil {
		t.Fatal("expected rand error")
	}
}

func TestGenerateGuessNumberRandIntError(t *testing.T) {
	orig := guessRandIntFn
	t.Cleanup(func() { guessRandIntFn = orig })
	guessRandIntFn = func(int, int) (int, error) {
		return 0, errors.New("rand int failed")
	}
	_, err := NewRandomNumberService().GenerateGuessNumber()
	if err == nil {
		t.Fatal("expected generate error")
	}
}

func TestRefreshTokenGenerateRandError(t *testing.T) {
	orig := refreshRandReadFn
	t.Cleanup(func() { refreshRandReadFn = orig })
	refreshRandReadFn = func([]byte) (int, error) {
		return 0, errors.New("rand read failed")
	}
	_, err := NewRefreshTokenService().Generate()
	if err == nil {
		t.Fatal("expected generate error")
	}
}

func TestJWTParseUnexpectedSigningMethod(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	claims := &accessClaims{
		Role: string(model.RoleUser),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.New().String(),
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString(svc.secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = svc.Parse(token)
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("parse: %v", err)
	}
}

func TestJWTParseInvalidTokenValidFlag(t *testing.T) {
	svc, err := NewJWTService("abcdefghijklmnopqrstuvwxyz123456", 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	orig := jwtParseWithClaimsFn
	t.Cleanup(func() { jwtParseWithClaimsFn = orig })
	jwtParseWithClaimsFn = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{Valid: false}, nil
	}
	_, err = svc.Parse("any-token")
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("parse: %v", err)
	}
}
