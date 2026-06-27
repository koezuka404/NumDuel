package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

func Login(ctx context.Context, d AuthDeps, in LoginInput) (*LoginOutput, error) {
	if err := domain.ValidateLoginEmail(in.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	user, err := d.Repo.Users().FindByEmailActive(ctx, in.Email)
	if err != nil {
		return nil, domain.ErrInternal("failed to find user")
	}
	if user == nil || !d.Passwords.Verify(user.PasswordHash, in.Password) {
		return nil, domain.ErrUnauthorized()
	}
	now := d.now()
	accessToken, err := d.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return nil, domain.ErrInternal("failed to issue access token")
	}
	refreshPair, err := d.RefreshTokens.Generate()
	if err != nil {
		return nil, domain.ErrInternal("failed to generate refresh token")
	}
	token := domain.NewRefreshToken(user.ID, refreshPair.Hash, uuid.New(), now.AddDate(0, 0, d.RefreshTokenExpiryDays), now)
	if err := withTx(ctx, d.Repo, func(tx domain.Transaction) error {
		if err := d.Repo.LoginLogs().Create(ctx, tx, &domain.LoginLog{
			ID: uuid.New(), UserID: user.ID, Action: domain.LoginActionLogin, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return domain.ErrInternal("failed to create login log")
		}
		user.LastActivityAt = now
		user.UpdatedAt = now
		if err := d.Repo.Users().Update(ctx, tx, user); err != nil {
			return domain.ErrInternal("failed to update user activity")
		}
		if err := d.Repo.RefreshTokens().Create(ctx, tx, &token); err != nil {
			return domain.ErrInternal("failed to store refresh token")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &LoginOutput{AccessToken: accessToken, RefreshToken: refreshPair.Plaintext}, nil
}
