package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

var refreshRandReadFn = rand.Read

var jwtParseWithClaimsFn = jwt.ParseWithClaims

type JWTService struct {
	secret        []byte
	expiryMinutes int
}

var _ usecase.IAccessTokenIssuer = (*JWTService)(nil)
var _ usecase.IAccessTokenParser = (*JWTService)(nil)

func NewJWTService(secret string, expiryMinutes int) (*JWTService, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 characters")
	}
	if expiryMinutes <= 0 {
		return nil, fmt.Errorf("JWT expiry must be positive")
	}
	return &JWTService{
		secret:        []byte(secret),
		expiryMinutes: expiryMinutes,
	}, nil
}

type accessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTService) Issue(userID uuid.UUID, role model.Role, now time.Time) (string, error) {
	claims := accessClaims{
		Role: string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.expiryMinutes) * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

type AccessToken struct {
	UserID    uuid.UUID
	Role      model.Role
	JTI       string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func (s *JWTService) Parse(tokenString string) (*usecase.AccessTokenClaims, error) {
	claims := &accessClaims{}
	token, err := jwtParseWithClaimsFn(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, usecase.ErrTokenExpired
		}
		return nil, usecase.ErrUnauthorized
	}
	if !token.Valid {
		return nil, usecase.ErrUnauthorized
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil || claims.ID == "" {
		return nil, usecase.ErrUnauthorized
	}
	out := &usecase.AccessTokenClaims{UserID: userID, Role: model.Role(claims.Role), JTI: claims.ID}
	if claims.IssuedAt != nil {
		out.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		out.ExpiresAt = claims.ExpiresAt.Time
	}
	return out, nil
}

type RefreshTokenService struct{}

var _ usecase.IRefreshTokenGenerator = (*RefreshTokenService)(nil)

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

func (s *RefreshTokenService) Hash(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func (s *RefreshTokenService) Generate() (usecase.RefreshTokenPair, error) {
	buf := make([]byte, 64)
	if _, err := refreshRandReadFn(buf); err != nil {
		return usecase.RefreshTokenPair{}, err
	}
	plaintext := hex.EncodeToString(buf)
	return usecase.RefreshTokenPair{Plaintext: plaintext, Hash: s.Hash(plaintext)}, nil
}
