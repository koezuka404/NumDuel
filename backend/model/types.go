// 列挙型・定数DB カラム値と API レスポンスで共通利用
package model

type GameStatus string

const (
	GameStatusWaitingSecret GameStatus = "WAITING_SECRET"
	GameStatusInProgress    GameStatus = "IN_PROGRESS"
	GameStatusFinished      GameStatus = "FINISHED"
)

type Role string

const (
	RoleUser   Role = "user"
	RoleMaster Role = "master"
)

type DigitResult int

const (
	DigitMiss DigitResult = 0
	DigitHit  DigitResult = 1
)

type RefreshTokenStatus string

const (
	RefreshTokenActive  RefreshTokenStatus = "active"
	RefreshTokenRevoked RefreshTokenStatus = "revoked"
)

type MatchingQueueStatus string

const (
	MatchingQueueWaiting MatchingQueueStatus = "waiting"
)

type LoginAction string

const (
	LoginActionLogin      LoginAction = "login"
	LoginActionLogout     LoginAction = "logout"
	LoginActionAutoLogout LoginAction = "auto_logout" // AutoLogoutWorker が記録
)
