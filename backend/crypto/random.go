package crypto

import (
	"crypto/rand"
	"math/big"

	"github.com/numduel/numduel/model"
)

type RandomNumberService struct{}

var _ model.GuessNumberGenerator = (*RandomNumberService)(nil)

func NewRandomNumberService() *RandomNumberService {
	return &RandomNumberService{}
}

// GenerateGuessNumber は重複なし 4 桁（0〜9 から 4 つ）を crypto/rand で生成する。
func (s *RandomNumberService) GenerateGuessNumber() (model.GuessNumber, error) {
	pool := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < 4; i++ {
		j, err := randInt(i, len(pool)-1)
		if err != nil {
			return model.GuessNumber{}, model.ErrInternal("failed to generate guess number")
		}
		pool[i], pool[j] = pool[j], pool[i]
	}
	return model.NewGuessNumber([4]int{pool[0], pool[1], pool[2], pool[3]})
}

func randInt(min, max int) (int, error) {
	if max < min {
		return 0, model.ErrInternal("invalid random range")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}
