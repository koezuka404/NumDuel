package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

// RegisterUserInput は RegisterUserUseCase の入力。
type RegisterUserInput struct {
	Username string
	Email    string
	Password string
}

// RegisterUserOutput は RegisterUserUseCase の出力（§6.3.1）。
type RegisterUserOutput struct {
	ID       uuid.UUID
	Username string
	Role     domain.Role
	WinCount int
}

// RegisterUserUseCase は新規ユーザーを登録する。
type RegisterUserUseCase struct {
	AuthDeps
}

func NewRegisterUserUseCase(deps AuthDeps) *RegisterUserUseCase {
	return &RegisterUserUseCase{AuthDeps: deps}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	if err := domain.ValidateUsername(input.Username); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	exists, err := uc.Repo.Users().ExistsByEmailOrUsername(ctx, input.Email, input.Username)
	if err != nil {
		return nil, domain.ErrInternal("failed to check duplicate user")
	}
	if exists {
		return nil, domain.ErrDuplicateUser()
	}

	hash, err := uc.Passwords.Hash(input.Password)
	if err != nil {
		return nil, domain.ErrInternal("failed to hash password")
	}

	now := uc.now()
	user := &domain.User{
		ID:             uuid.New(),
		Username:       input.Username,
		Email:          input.Email,
		PasswordHash:   hash,
		Role:           domain.RoleUser,
		WinCount:       0,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	tx, err := uc.Repo.Begin(ctx)
	if err != nil {
		return nil, domain.ErrInternal("failed to begin transaction")
	}
	defer func() { _ = uc.Repo.Rollback(tx) }()

	if err := uc.Repo.Users().Create(ctx, tx, user); err != nil {
		return nil, domain.ErrInternal("failed to create user")
	}
	if err := uc.Repo.Commit(tx); err != nil {
		return nil, domain.ErrInternal("failed to commit transaction")
	}

	return &RegisterUserOutput{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		WinCount: user.WinCount,
	}, nil
}
