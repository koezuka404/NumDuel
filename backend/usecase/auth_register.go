package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

func (a *AuthUseCase) Register(ctx context.Context, in RegisterInput) (*LoginResult, error) {
	if err := ValidateUsername(in.Username); err != nil {
		return nil, mapValidationErr(err)
	}
	if err := ValidateEmail(in.Email); err != nil {
		return nil, mapValidationErr(err)
	}
	if err := ValidatePassword(in.Password); err != nil {
		return nil, mapValidationErr(err)
	}
	exists, err := emailOrUsernameExists(ctx, a.Users, in.Email, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUser
	}
	hash, err := a.Passwords.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	now := a.now()
	user := &model.User{
		ID:             uuid.New(),
		Username:       in.Username,
		Email:          in.Email,
		PasswordHash:   hash,
		Role:           model.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	var tokens *SessionTokens
	if err := repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		if err := a.Users.Create(ctx, user); err != nil {
			return err
		}
		if err := a.createLoginLog(ctx, user.ID, model.LoginActionLogin, now); err != nil {
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
