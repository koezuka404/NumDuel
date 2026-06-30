package usecase

import "github.com/numduel/numduel/model"

func JudgeDigits(secret, guess [4]int) [4]model.DigitResult {
	var results [4]model.DigitResult
	for i := 0; i < 4; i++ {
		if secret[i] == guess[i] {
			results[i] = model.DigitHit
		}
	}
	return results
}

func IsWin(results [4]model.DigitResult) bool {
	for _, r := range results {
		if r != model.DigitHit {
			return false
		}
	}
	return true
}

func HitCount(results [4]model.DigitResult) int {
	n := 0
	for _, r := range results {
		if r == model.DigitHit {
			n++
		}
	}
	return n
}

func DigitResultsToInts(results []model.DigitResult) []int {
	out := make([]int, len(results))
	for i, r := range results {
		out[i] = int(r)
	}
	return out
}
