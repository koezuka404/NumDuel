package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

// LoginInput は LoginUseCase の入力。
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput は LoginUseCase の出力。
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

// LoginUseCase はログイン認証を行う。
type LoginUseCase struct {
	AuthDeps
}

func NewLoginUseCase(deps AuthDeps) *LoginUseCase {
	return &LoginUseCase{AuthDeps: deps}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := domain.ValidateLoginEmail(input.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	user, err := uc.Repo.Users().FindByEmailActive(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInternal("failed to find user")
	}
	if user == nil || !uc.Passwords.Verify(user.PasswordHash, input.Password) {
		return nil, domain.ErrUnauthorized()
	}

	now := uc.now()
	accessToken, _, err := uc.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return nil, domain.ErrInternal("failed to issue access token")
	}

	refreshPair, err := uc.RefreshTokens.Generate()
	if err != nil {
		return nil, domain.ErrInternal("failed to generate refresh token")
	}

	familyID := uuid.New()
	expiresAt := now.AddDate(0, 0, uc.RefreshTokenExpiryDays)
	token := domain.NewRefreshToken(user.ID, refreshPair.Hash, familyID, expiresAt, now)

	tx, err := uc.Repo.Begin(ctx)
	if err != nil {
		return nil, domain.ErrInternal("failed to begin transaction")
	}
	defer func() { _ = uc.Repo.Rollback(tx) }()

	loginLog := &domain.LoginLog{
		ID:        uuid.New(),
		UserID:    user.ID,
		Action:    domain.LoginActionLogin,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.Repo.LoginLogs().Create(ctx, tx, loginLog); err != nil {
		return nil, domain.ErrInternal("failed to create login log")
	}

	user.LastActivityAt = now
	user.UpdatedAt = now
	if err := uc.Repo.Users().Update(ctx, tx, user); err != nil {
		return nil, domain.ErrInternal("failed to update user activity")
	}

	if err := uc.Repo.RefreshTokens().Create(ctx, tx, &token); err != nil {
		return nil, domain.ErrInternal("failed to store refresh token")
	}

	if err := uc.Repo.Commit(tx); err != nil {
		return nil, domain.ErrInternal("failed to commit transaction")
	}

	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshPair.Plaintext,
	}, nil
}
