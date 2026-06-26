package domain

// 桁判定・勝敗判定の純粋関数（仕様書 3.5）。
//
// 配置ルール（Clean Architecture）:
//   - Controller / WebSocket / Infrastructure 内で判定してはならない
//   - DB / Redis / WebSocket を呼び出さない（副作用ゼロ）
//   - 単体テストカバレッジ 100% 必須（仕様書 18.4）
//
// 本番の予想照合は Infrastructure.SecretHashService.Verify が
// DB 上の HMAC ハッシュと照合するが、結果は JudgeDigits と同一であること。

// JudgeDigits は秘密数字と予想数字を桁ごとに比較する（仕様書 3.5, 1.3）。
//
// 判定ルール:
//   - 各桁 i について secret[i] == guess[i] なら DigitHit (1)
//   - それ以外は DigitMiss (0)
//   - 「数字は含むが位置が違う」は外れ（Bulls & Cows の Cow 相当は採用しない）
//
// 例（仕様書 3.5）:
//
//	secret=1234, guess=1234 → [1,1,1,1], hitCount=4
//	secret=1234, guess=1456 → [0,1,0,0], hitCount=1
//	secret=1234, guess=5678 → [0,0,0,0], hitCount=0
//	secret=1234, guess=4321 → [0,0,0,0], hitCount=0（位置不一致はすべて外れ）
func JudgeDigits(secret, guess [4]int) [4]DigitResult {
	var results [4]DigitResult
	for i := 0; i < 4; i++ {
		if secret[i] == guess[i] {
			results[i] = DigitHit
		} else {
			results[i] = DigitMiss
		}
	}
	return results
}

// IsWin は 4 桁すべてが当たりか判定する（仕様書 3.5, 1.5）。
//
// hitCount == 4 と等価。SubmitGuessUseCase は Verify 結果に対して呼び、
// true なら同一 TX 内で FinishGameService を実行する。
func IsWin(results [4]DigitResult) bool {
	for _, r := range results {
		if r != DigitHit {
			return false
		}
	}
	return true
}
