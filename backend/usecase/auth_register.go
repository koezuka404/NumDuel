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

func (a *AuthUseCase) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
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
	if err := repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		return a.Users.Create(ctx, user)
	}); err != nil {
		return nil, err
	}
	return &RegisterResult{
		ID:       user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
		WinCount: user.WinCount,
	}, nil
}
