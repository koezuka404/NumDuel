package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string // Controller が Cookie にセット
	RefreshToken string
	ID           uuid.UUID
	Username     string
	Role         model.Role
}

// Login は認証し、JWT と refresh_token（DB にはハッシュのみ保存）を発行する
func Login(ctx context.Context, d AuthDeps, in LoginInput) (*LoginOutput, error) {
	if err := model.ValidateLoginEmail(in.Email); err != nil {
		return nil, err
	}
	if err := model.ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	user, err := findUserByEmailActive(ctx, d.Repo, in.Email)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || !d.Passwords.Verify(user.PasswordHash, in.Password) {
		return nil, model.ErrUnauthorized()
	}
	now := d.now()
	accessToken, err := d.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return nil, model.ErrInternal("failed to issue access token")
	}
	refreshPair, err := d.RefreshTokens.Generate()
	if err != nil {
		return nil, model.ErrInternal("failed to generate refresh token")
	}
	// family_id は新規 UUID同一ログインセッション内のローテーションで共有
	token := model.NewRefreshToken(user.ID, refreshPair.Hash, uuid.New(), now.AddDate(0, 0, d.RefreshTokenExpiryDays), now)
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		if err := tx.LoginLogs().Create(ctx, &model.LoginLog{
			ID: uuid.New(), UserID: user.ID, Action: model.LoginActionLogin, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return model.ErrInternal("failed to create login log")
		}
		user.LastActivityAt = now
		user.UpdatedAt = now
		if err := tx.Users().Update(ctx, user); err != nil {
			return model.ErrInternal("failed to update user activity")
		}
		if err := tx.RefreshTokens().Create(ctx, &token); err != nil {
			return model.ErrInternal("failed to store refresh token")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &LoginOutput{
		AccessToken: accessToken, RefreshToken: refreshPair.Plaintext,
		ID: user.ID, Username: user.Username, Role: user.Role,
	}, nil
}
