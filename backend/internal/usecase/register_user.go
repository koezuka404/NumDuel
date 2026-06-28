package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

type RegisterUserInput struct {
	Username string
	Email    string
	Password string
}

type RegisterUserOutput struct {
	ID       uuid.UUID
	Username string
	Role     domain.Role
	WinCount int
}

// RegisterUser は新規ユーザーを登録する。JWT / refresh / login_logs は作成しない。
func RegisterUser(ctx context.Context, d AuthDeps, in RegisterUserInput) (*RegisterUserOutput, error) {
	if err := domain.ValidateUsername(in.Username); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(in.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	exists, err := d.Repo.Users().ExistsByEmailOrUsername(ctx, in.Email, in.Username)
	if err != nil {
		return nil, domain.ErrInternal("failed to check duplicate user")
	}
	if exists {
		return nil, domain.ErrDuplicateUser()
	}
	hash, err := d.Passwords.Hash(in.Password)
	if err != nil {
		return nil, domain.ErrInternal("failed to hash password")
	}
	now := d.now()
	user := &domain.User{
		ID:             uuid.New(),
		Username:       in.Username,
		Email:          in.Email,
		PasswordHash:   hash,
		Role:           domain.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := withTx(ctx, d.Repo, func(tx domain.Transaction) error {
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
