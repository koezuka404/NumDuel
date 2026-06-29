package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

type SecretHashService struct {
	pepper []byte
}

var _ model.SecretHasher = (*SecretHashService)(nil)

func NewSecretHashService(pepper string) (*SecretHashService, error) {
	if len(pepper) < 32 {
		return nil, fmt.Errorf("GAME_SECRET_PEPPER must be at least 32 bytes")
	}
	return &SecretHashService{pepper: []byte(pepper)}, nil
}

func (s *SecretHashService) Hash(digits [4]int, gameID uuid.UUID, playerSlot int) (string, error) {
	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		parts[i] = s.digest(gameID, playerSlot, i, digits[i])
	}
	return strings.Join(parts, ":"), nil
}

func (s *SecretHashService) Verify(storedHash string, guess model.GuessNumber, gameID uuid.UUID, opponentSlot int) ([4]model.DigitResult, error) {
	parts := strings.Split(storedHash, ":")
	if len(parts) != 4 {
		return [4]model.DigitResult{}, model.ErrInternal("invalid secret hash format")
	}
	digits := guess.Digits()
	var results [4]model.DigitResult
	for i := 0; i < 4; i++ {
		expected := s.digest(gameID, opponentSlot, i, digits[i])
		if hmac.Equal([]byte(parts[i]), []byte(expected)) {
			results[i] = model.DigitHit
		}
	}
	return results, nil
}

func (s *SecretHashService) digest(gameID uuid.UUID, playerSlot, position, digit int) string {
	msg := fmt.Sprintf("%s:%d:%d:%d", gameID.String(), playerSlot, position, digit)
	mac := hmac.New(sha256.New, s.pepper)
	_, _ = mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}
