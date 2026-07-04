package crypto

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/numduel/numduel/usecase"
)

const bcryptCost = 12

var generateFromPasswordFn = bcrypt.GenerateFromPassword

type PasswordService struct{}

var _ usecase.IPasswordHasher = (*PasswordService)(nil)

func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

func (s *PasswordService) Hash(password string) (string, error) {
	b, err := generateFromPasswordFn([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *PasswordService) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
