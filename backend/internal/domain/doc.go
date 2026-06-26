// Package domain は Num Duel のドメイン層（仕様書 第4章）を定義する。
//
// # 設計原則（仕様書 3.2〜3.5）
//
//   - ゲームルール・状態遷移の真実は Domain に置く
//   - DB / Redis / WebSocket へのアクセスは禁止（副作用なし）
//   - 桁判定（JudgeDigits / IsWin）は純粋関数のみ
//   - ゲーム状態の変更は Entity のメソッド経由（UseCase が TX 内で呼び出す）
//
// # ファイル構成（仕様書 第17章）
//
//   - types.go         … 値オブジェクト用の列挙型（GameStatus, Role 等）
//   - errors.go        … ドメインエラー（HTTP マッピングは UseCase / Controller 側）
//   - user.go          … User エンティティ
//   - guess.go         … SecretNumber / GuessNumber（値オブジェクト）と Guess エンティティ
//   - game.go          … Game 集約ルート
//   - ranking.go       … MatchHistory / Ranking（読み取りモデル）
//   - refresh_token.go … RefreshToken エンティティ
//   - judge.go         … 桁判定・勝敗判定の純粋関数
package domain
