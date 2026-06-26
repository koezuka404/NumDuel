package domain

// GameStatus は対戦ゲームのライフサイクル状態（仕様書 4.1, 9.4）。
//
// 遷移の概要:
//
//	WAITING_SECRET … マッチング成立後、両者の秘密数字登録待ち
//	IN_PROGRESS    … 両者登録完了後、交互ターン制の対戦中
//	FINISHED       … 勝敗確定（guess_win）または期限切れ終了（secret_setup_timeout）
type GameStatus string

const (
	// GameStatusWaitingSecret は秘密数字登録フェーズ。
	// SECRET_SETUP_SECONDS 以内に両者が登録しない場合 CancelGameBySecretTimeout により FINISHED へ。
	GameStatusWaitingSecret GameStatus = "WAITING_SECRET"

	// GameStatusInProgress は対戦中。SubmitGuess / HandleTimeout が有効。
	GameStatusInProgress GameStatus = "IN_PROGRESS"

	// GameStatusFinished は終了済み。これ以降の GUESS / SET_SECRET は不可。
	GameStatusFinished GameStatus = "FINISHED"
)

// Role はユーザー権限（仕様書 4.1, 9.3）。
//
//   - RoleUser   … 通常ユーザー。対戦・マッチング可能
//   - RoleMaster … 管理者。管理 API のみ。マッチング・対戦不可（仕様書 5.x, 13.13）
type Role string

const (
	RoleUser   Role = "user"
	RoleMaster Role = "master"
)

// DigitResult は各桁の判定結果（仕様書 1.3, 4.1）。
//
// API / WS では整数 0/1 で表現し、フロントエンドが ○/× に変換する。
//  Bulls & Cows の「数字は含むが位置が違う」は当たりにしない（位置一致のみ）。
type DigitResult int

const (
	// DigitMiss (0) … 外れ。その桁は一致しない。
	DigitMiss DigitResult = 0

	// DigitHit (1) … 当たり。位置と数字が両方一致。
	DigitHit DigitResult = 1
)

// RefreshTokenStatus はリフレッシュトークンの状態（仕様書 4.1, Login/Refresh UseCase）。
type RefreshTokenStatus string

const (
	// RefreshTokenActive … 有効。RefreshTokenUseCase で検証可能。
	RefreshTokenActive RefreshTokenStatus = "active"

	// RefreshTokenRevoked … 失効。ログアウト・ローテーション・盗用検出で設定。
	RefreshTokenRevoked RefreshTokenStatus = "revoked"
)
