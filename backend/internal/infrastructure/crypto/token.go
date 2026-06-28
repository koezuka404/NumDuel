// パスワードハッシュ（bcrypt）と JWT / refresh トークンの生成・検証。
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

	"github.com/numduel/numduel/internal/domain"
)

type JWTService struct {
	secret        []byte
	expiryMinutes int
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

// Issue は accessToken（JWT）を発行する。Claims: sub, role, jti, iat, exp
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

// AccessToken は Parse の結果。Middleware / Logout で使用。
type AccessToken struct {
	UserID    uuid.UUID
	Role      domain.Role
	JTI       string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Parse は JWT を検証して Claims を取り出す。期限切れは token_expired を返す。
func (s *JWTService) Parse(tokenString string) (*AccessToken, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired()
		}
		return nil, domain.ErrUnauthorized()
	}
	if !token.Valid {
		return nil, domain.ErrUnauthorized()
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil || claims.ID == "" {
		return nil, domain.ErrUnauthorized()
	}
	var issuedAt, expiresAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	return &AccessToken{
		UserID: userID, Role: domain.Role(claims.Role), JTI: claims.ID,
		IssuedAt: issuedAt, ExpiresAt: expiresAt,
	}, nil
}

type RefreshTokenService struct{}

var _ domain.RefreshTokenGenerator = (*RefreshTokenService)(nil)

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

// HashRefreshToken は平文 refresh を SHA-256 hex に変換（DB 照合用）。
func HashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// Generate は 64 バイト乱数 → hex 平文 + ハッシュのペアを生成。
func (s *RefreshTokenService) Generate() (domain.RefreshTokenPair, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return domain.RefreshTokenPair{}, err
	}
	plaintext := hex.EncodeToString(buf)
	return domain.RefreshTokenPair{
		Plaintext: plaintext,
		Hash:      HashRefreshToken(plaintext),
	}, nil
}
