package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

type AccessTokenIssuer interface {
	Issue(userID uuid.UUID, role Role, now time.Time) (string, error)
}

type RefreshTokenPair struct {
	Plaintext string
	Hash      string
}

type RefreshTokenGenerator interface {
	Generate() (RefreshTokenPair, error)
}
