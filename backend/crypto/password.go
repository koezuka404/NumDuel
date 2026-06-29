// bcrypt によるパスワードハッシュ・照合。
package crypto

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/numduel/numduel/model"
)

const bcryptCost = 12

type PasswordService struct{}

var _ model.PasswordHasher = (*PasswordService)(nil)

func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

func (s *PasswordService) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *PasswordService) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
