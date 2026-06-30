package usecase

import (
	"context"

	"github.com/google/uuid"
)

func (a *AuthUseCase) GetMe(ctx context.Context, userID uuid.UUID) (*MeResult, error) {
	user, err := a.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted() {
		return nil, ErrUnauthorized
	}
	return &MeResult{
		ID:       user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
		WinCount: user.WinCount,
	}, nil
}
