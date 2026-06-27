package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

type JWTService struct {
	secret         []byte
	expiryMinutes  int
}

var _ domain.AccessTokenIssuer = (*JWTService)(nil)

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

func (s *JWTService) Issue(userID uuid.UUID, role domain.Role, now time.Time) (string, error) {
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

type RefreshTokenService struct{}

var _ domain.RefreshTokenGenerator = (*RefreshTokenService)(nil)

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

func (s *RefreshTokenService) Generate() (domain.RefreshTokenPair, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return domain.RefreshTokenPair{}, err
	}
	plaintext := hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plaintext))
	return domain.RefreshTokenPair{
		Plaintext: plaintext,
		Hash:      hex.EncodeToString(sum[:]),
	}, nil
}
