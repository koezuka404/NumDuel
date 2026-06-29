package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

type RegisterUserInput struct {
	Username string
	Email    string
	Password string
}

type RegisterUserOutput struct {
	ID       uuid.UUID
	Username string
	Role     model.Role
	WinCount int
}

// RegisterUser は新規ユーザーを登録する。JWT / refresh / login_logs は作成しない。
func RegisterUser(ctx context.Context, d AuthDeps, in RegisterUserInput) (*RegisterUserOutput, error) {
	if err := model.ValidateUsername(in.Username); err != nil {
		return nil, err
	}
	if err := model.ValidateEmail(in.Email); err != nil {
		return nil, err
	}
	if err := model.ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	exists, err := emailOrUsernameExists(ctx, d.Repo, in.Email, in.Username)
	if err != nil {
		return nil, model.ErrInternal("failed to check duplicate user")
	}
	if exists {
		return nil, model.ErrDuplicateUser()
	}
	hash, err := d.Passwords.Hash(in.Password)
	if err != nil {
		return nil, model.ErrInternal("failed to hash password")
	}
	now := d.now()
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
	if err := withTx(ctx, d.Repo, func(tx model.Transaction) error {
		return d.Repo.Users().Create(ctx, tx, user)
	}); err != nil {
		return nil, err
	}
	return &RegisterUserOutput{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		WinCount: user.WinCount,
	}, nil
}
