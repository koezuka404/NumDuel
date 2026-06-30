package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

const masterUsername = "admin"

// SeedMasterInput は初回 master 作成用
type SeedMasterInput struct {
	Email    string
	Password string
}

// SeedMaster は master が未存在かつ env が設定されているとき 1 件だけ作成する
func SeedMaster(ctx context.Context, d AuthDeps, in SeedMasterInput) error {
	if in.Email == "" || in.Password == "" {
		return nil
	}
	exists, err := d.Repo.User.ExistsActiveMaster(ctx)
	if err != nil {
		return model.ErrInternal("failed to check master user")
	}
	if exists {
		return nil
	}
	if err := ValidateLoginEmail(in.Email); err != nil {
		return err
	}
	if err := ValidatePassword(in.Password); err != nil {
		return err
	}
	dup, err := emailOrUsernameExists(ctx, d.Repo, in.Email, masterUsername)
	if err != nil {
		return model.ErrInternal("failed to check duplicate user")
	}
	if dup {
		return model.ErrInternal("master seed email or username already taken")
	}
	hash, err := d.Passwords.Hash(in.Password)
	if err != nil {
		return model.ErrInternal("failed to hash password")
	}
	now := d.now()
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
	return repository.WithTx(ctx, d.Repo.DB, func(ctx context.Context) error {
		return d.Repo.User.Create(ctx, user)
	})
}
