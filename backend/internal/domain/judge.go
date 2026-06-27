package domain

func JudgeDigits(secret, guess [4]int) [4]DigitResult {
	var results [4]DigitResult
	for i := 0; i < 4; i++ {
		if secret[i] == guess[i] {
			results[i] = DigitHit
		}
	}
	return results
}

func IsWin(results [4]DigitResult) bool {
	for _, r := range results {
		if r != DigitHit {
			return false
		}
	}
	return true
}
