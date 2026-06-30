package model

type GameStatus string

const (
	GameStatusWaitingSecret GameStatus = "WAITING_SECRET"
	GameStatusInProgress    GameStatus = "IN_PROGRESS"
	GameStatusFinished      GameStatus = "FINISHED"
)

func (s GameStatus) Valid() bool {
	switch s {
	case GameStatusWaitingSecret, GameStatusInProgress, GameStatusFinished:
		return true
	default:
		return false
	}
}

type Role string

const (
	RoleUser   Role = "user"
	RoleMaster Role = "master"
)

func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleMaster:
		return true
	default:
		return false
	}
}

type DigitResult int

const (
	DigitMiss DigitResult = 0
	DigitHit  DigitResult = 1
)

func (d DigitResult) Valid() bool {
	return d == DigitMiss || d == DigitHit
}

type RefreshTokenStatus string

const (
	RefreshTokenActive  RefreshTokenStatus = "active"
	RefreshTokenRevoked RefreshTokenStatus = "revoked"
)

func (s RefreshTokenStatus) Valid() bool {
	return s == RefreshTokenActive || s == RefreshTokenRevoked
}

type MatchingQueueStatus string

const (
	MatchingQueueWaiting MatchingQueueStatus = "waiting"
)

func (s MatchingQueueStatus) Valid() bool {
	return s == MatchingQueueWaiting
}

type LoginAction string

const (
	LoginActionLogin      LoginAction = "login"
	LoginActionLogout     LoginAction = "logout"
	LoginActionAutoLogout LoginAction = "auto_logout"
)

func (a LoginAction) Valid() bool {
	switch a {
	case LoginActionLogin, LoginActionLogout, LoginActionAutoLogout:
		return true
	default:
		return false
	}
}
