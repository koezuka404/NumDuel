package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func (a *AuthUseCase) issueSession(ctx context.Context, user *model.User) (*SessionTokens, error) {
	now := a.now()
	accessToken, err := a.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return nil, err
	}
	pair, err := a.RefreshGen.Generate()
	if err != nil {
		return nil, err
	}
	token := model.NewRefreshToken(
		user.ID,
		pair.Hash,
		uuid.New(),
		tokenExpiry(a, now),
		now,
	)
	if err := a.RefreshTokens.Create(ctx, &token); err != nil {
		return nil, err
	}
	return &SessionTokens{AccessToken: accessToken, RefreshToken: pair.Plaintext}, nil
}

func tokenExpiry(a *AuthUseCase, now time.Time) time.Time {
	return now.AddDate(0, 0, a.RefreshDays)
}

func (a *AuthUseCase) touchUserActivity(ctx context.Context, user *model.User, now time.Time) error {
	user.LastActivityAt = now
	user.UpdatedAt = now
	return a.Users.Update(ctx, user)
}

func (a *AuthUseCase) createLoginLog(ctx context.Context, userID uuid.UUID, action model.LoginAction, now time.Time) error {
	return a.LoginLogs.Create(ctx, &model.LoginLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
