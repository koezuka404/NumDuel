// 桁判定の純粋関数。DB/HTTP に依存しない（Domain 層の核心ロジック）。
package domain

// JudgeDigits は各桁を secret と比較し 0=miss / 1=hit を返す。
func JudgeDigits(secret, guess [4]int) [4]DigitResult {
	var results [4]DigitResult
	for i := 0; i < 4; i++ {
		if secret[i] == guess[i] {
			results[i] = DigitHit
		}
	}
	return results
}

// IsWin は 4 桁すべて hit なら true。
func IsWin(results [4]DigitResult) bool {
	for _, r := range results {
		if r != DigitHit {
			return false
		}
	}
	return true
}
