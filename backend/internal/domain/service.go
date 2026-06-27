package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordHasher はパスワードのハッシュ化と照合。
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// AccessTokenClaims は JWT に載せるクレーム。
type AccessTokenClaims struct {
	UserID uuid.UUID
	Role   Role
	JTI    string
	Issued time.Time
	Expiry time.Time
}

// AccessTokenIssuer は JWT アクセストークンを発行する。
type AccessTokenIssuer interface {
	Issue(userID uuid.UUID, role Role, now time.Time) (token string, claims AccessTokenClaims, err error)
}

// RefreshTokenPair は平文トークンと DB 保存用ハッシュ。
type RefreshTokenPair struct {
	Plaintext string
	Hash      string
}

// RefreshTokenGenerator はリフレッシュトークンを生成する。
type RefreshTokenGenerator interface {
	Generate() (RefreshTokenPair, error)
}
