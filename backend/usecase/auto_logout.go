package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

//非アクティブユーザーの自動ログアウトユースケース。
type IAutoLogoutUsecase interface {
	Run(ctx context.Context) error
}

type AutoLogoutUseCase struct {
	Users           repository.IUserRepo
	RefreshTokens   repository.IRefreshTokenRepo
	LoginLogs       repository.ILoginLogRepo
	DB              *gorm.DB
	ForceLogout     IForceLogoutStore
	ForceDisconnect func(ctx context.Context, userID uuid.UUID) error
	SessionTimeout  time.Duration
	Now             func() time.Time
}

func (a *AutoLogoutUseCase) now() time.Time {
	if a != nil && a.Now != nil {
		return a.Now().UTC()
	}
	return time.Now().UTC()
}

func (a *AutoLogoutUseCase) Run(ctx context.Context) error {
	if a.SessionTimeout <= 0 {
		return nil
	}
	now := a.now()
	before := now.Add(-a.SessionTimeout)
	users, err := a.Users.ListInactiveSince(ctx, before)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == nil {
			continue
		}
		if err := a.logoutUser(ctx, user.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func (a *AutoLogoutUseCase) logoutUser(ctx context.Context, userID uuid.UUID, now time.Time) error {
	if a.ForceLogout != nil {
		if err := a.ForceLogout.SetForceLogoutBefore(ctx, userID, now); err != nil {
			return err
		}
	}
	if a.ForceDisconnect != nil {
		_ = a.ForceDisconnect(ctx, userID)
	}
	return repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		if err := revokeRefreshTokensByUserID(ctx, a.RefreshTokens, userID, now); err != nil {
			return err
		}
		if err := a.LoginLogs.Create(ctx, &model.LoginLog{
			ID: uuid.New(), UserID: userID, Action: model.LoginActionAutoLogout,
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return err
		}
		user, err := a.Users.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user != nil && !user.IsDeleted() {
			user.LastActivityAt = now
			user.UpdatedAt = now
			return a.Users.Update(ctx, user)
		}
		return nil
	})
}

func NewAutoLogoutUseCase(repos repository.Repos, forceLogout IForceLogoutStore, forceDisconnect func(context.Context, uuid.UUID) error, sessionTimeout time.Duration) *AutoLogoutUseCase {
	return &AutoLogoutUseCase{
		Users:           repos.User,
		RefreshTokens:   repos.RefreshToken,
		LoginLogs:       repos.LoginLog,
		DB:              repos.DB,
		ForceLogout:     forceLogout,
		ForceDisconnect: forceDisconnect,
		SessionTimeout:  sessionTimeout,
	}
}
