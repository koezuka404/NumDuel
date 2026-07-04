package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/numduel/numduel/usecase"
)

type RandomNumberService struct{}

var _ usecase.IGuessNumberGenerator = (*RandomNumberService)(nil)

func NewRandomNumberService() *RandomNumberService {
	return &RandomNumberService{}
}

var (
	guessRandIntFn = randInt
	intRandFn      = rand.Int
)

func (s *RandomNumberService) GenerateGuessNumber() (string, error) {
	pool := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < 4; i++ {
		j, err := guessRandIntFn(i, len(pool)-1)
		if err != nil {
			return "", err
		}
		pool[i], pool[j] = pool[j], pool[i]
	}
	return fmt.Sprintf("%d%d%d%d", pool[0], pool[1], pool[2], pool[3]), nil
}

func randInt(min, max int) (int, error) {
	if max < min {
		return 0, fmt.Errorf("invalid random range")
	}
	n, err := intRandFn(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}
