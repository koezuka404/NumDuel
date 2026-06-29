package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

type GetMeOutput struct {
	ID       uuid.UUID
	Username string
	Role     model.Role
	WinCount int
}

// GetMe は JWT で特定されたユーザーの基本情報を返す。
func GetMe(ctx context.Context, d AuthDeps, userID uuid.UUID) (*GetMeOutput, error) {
	user, err := d.Repo.Users().FindByID(ctx, userID)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		return nil, model.ErrUnauthorized()
	}
	return &GetMeOutput{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		WinCount: user.WinCount,
	}, nil
}
