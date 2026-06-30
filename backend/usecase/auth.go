package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"gorm.io/gorm"
)

//Register/Login/Refresh/Logout/Meを扱う認証ユースケース。
type IAuthUsecase interface {
	Register(ctx context.Context, in RegisterInput) (*RegisterResult, error)
	Login(ctx context.Context, in LoginInput) (*LoginResult, error)
	Refresh(ctx context.Context, in RefreshInput) (*RefreshResult, error)
	Logout(ctx context.Context, in LogoutInput) error
	GetMe(ctx context.Context, userID uuid.UUID) (*MeResult, error)
	SeedMaster(ctx context.Context, in SeedMasterInput) error
	CleanupExpiredRefreshTokens(ctx context.Context)
}

//パスワードのhash化と照合。
type IPasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

//AccessToken発行。
type IAccessTokenIssuer interface {
	Issue(userID uuid.UUID, role model.Role, now time.Time) (string, error)
}

type RefreshTokenPair struct {
	Plaintext string
	Hash      string
}

//RefreshToken生成とhash化。
type IRefreshTokenGenerator interface {
	Generate() (RefreshTokenPair, error)
	Hash(plaintext string) string
}

type AccessTokenClaims struct {
	UserID    uuid.UUID
	Role      model.Role
	JTI       string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

//JWT失効管理。
type IJWTRevoker interface {
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

//WebSocketセッションのRedis管理。
type IWSSessionStore interface {
	SetUser(ctx context.Context, userID uuid.UUID, connectionID string, ttl time.Duration) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

//強制ログアウト時刻のRedis管理。
type IForceLogoutStore interface {
	SetForceLogoutBefore(ctx context.Context, userID uuid.UUID, at time.Time) error
	GetForceLogoutBefore(ctx context.Context, userID uuid.UUID) (time.Time, error)
}

const defaultRefreshTokenCleanupGraceDays = 7

type SessionTokens struct {
	AccessToken  string
	RefreshToken string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ID           string
	Username     string
	Role         string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type RegisterResult struct {
	ID       string
	Username string
	Role     string
	WinCount int
}

type MeResult struct {
	ID       string
	Username string
	Role     string
	WinCount int
}

type AuthUseCase struct {
	Users         repository.IUserRepo
	RefreshTokens repository.IRefreshTokenRepo
	LoginLogs     repository.ILoginLogRepo
	DB            *gorm.DB
	Passwords     IPasswordHasher
	AccessTokens  IAccessTokenIssuer
	RefreshGen    IRefreshTokenGenerator
	JWTRevoker    IJWTRevoker
	WSSessions    IWSSessionStore
	RefreshDays   int
	CleanupGrace  int
	Now           func() time.Time
}

func (a *AuthUseCase) now() time.Time {
	if a != nil && a.Now != nil {
		return a.Now().UTC()
	}
	return time.Now().UTC()
}

func (a *AuthUseCase) cleanupGraceDays() int {
	if a == nil || a.CleanupGrace <= 0 {
		return defaultRefreshTokenCleanupGraceDays
	}
	return a.CleanupGrace
}

func NewAuthUseCase(repos repository.Repos, passwords IPasswordHasher, access IAccessTokenIssuer, refresh IRefreshTokenGenerator, jwtRevoker IJWTRevoker, ws IWSSessionStore, refreshDays, cleanupGrace int) *AuthUseCase {
	return &AuthUseCase{
		Users:         repos.User,
		RefreshTokens: repos.RefreshToken,
		LoginLogs:     repos.LoginLog,
		DB:            repos.DB,
		Passwords:     passwords,
		AccessTokens:  access,
		RefreshGen:    refresh,
		JWTRevoker:    jwtRevoker,
		WSSessions:    ws,
		RefreshDays:   refreshDays,
		CleanupGrace:  cleanupGrace,
	}
}
