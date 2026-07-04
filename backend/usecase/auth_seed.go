package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

const masterUsername = "admin"

type SeedMasterInput struct {
	Email    string
	Password string
}

func (a *AuthUseCase) SeedMaster(ctx context.Context, in SeedMasterInput) error {
	if in.Email == "" || in.Password == "" {
		return nil
	}
	exists, err := a.Users.ExistsActiveMaster(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := ValidateLoginEmail(in.Email); err != nil {
		return mapValidationErr(err)
	}
	if err := ValidatePassword(in.Password); err != nil {
		return mapValidationErr(err)
	}
	dup, err := emailOrUsernameExists(ctx, a.Users, in.Email, masterUsername)
	if err != nil {
		return err
	}
	if dup {
		return errors.New("管理者のメールまたはユーザー名が既に使用されています")
	}
	hash, err := a.Passwords.Hash(in.Password)
	if err != nil {
		return err
	}
	now := a.now()
	user := &model.User{
		ID:             uuid.New(),
		Username:       masterUsername,
		Email:          in.Email,
		PasswordHash:   hash,
		Role:           model.RoleMaster,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		return a.Users.Create(ctx, user)
	})
}
