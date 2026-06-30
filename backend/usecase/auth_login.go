package usecase

import (
	"context"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type LoginInput struct {
	Email    string
	Password string
}

func (a *AuthUseCase) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if err := ValidateLoginEmail(in.Email); err != nil {
		return nil, mapValidationErr(err)
	}
	if err := ValidatePassword(in.Password); err != nil {
		return nil, mapValidationErr(err)
	}
	user, err := findUserByEmailActive(ctx, a.Users, in.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || !a.Passwords.Verify(user.PasswordHash, in.Password) {
		return nil, ErrUnauthorized
	}
	now := a.now()
	var tokens *SessionTokens
	if err := repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		if err := a.createLoginLog(ctx, user.ID, model.LoginActionLogin, now); err != nil {
			return err
		}
		if err := a.touchUserActivity(ctx, user, now); err != nil {
			return err
		}
		var err error
		tokens, err = a.issueSession(ctx, user)
		return err
	}); err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ID:           user.ID.String(),
		Username:     user.Username,
		Role:         string(user.Role),
	}, nil
}
