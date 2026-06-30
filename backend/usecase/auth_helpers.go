package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func findUserByEmailActive(ctx context.Context, users repository.IUserRepo, email string) (*model.User, error) {
	user, err := users.FindByEmail(ctx, email)
	if err != nil || user == nil || user.IsDeleted() {
		return nil, err
	}
	return user, nil
}

func emailOrUsernameExists(ctx context.Context, users repository.IUserRepo, email, username string) (bool, error) {
	byEmail, err := users.FindByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if byEmail != nil {
		return true, nil
	}
	byUsername, err := users.FindByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return byUsername != nil, nil
}

func revokeRefreshTokensByUserID(ctx context.Context, repo repository.IRefreshTokenRepo, userID uuid.UUID, now time.Time) error {
	return repo.RevokeByUserID(ctx, userID, now)
}

func revokeRefreshTokenFamily(ctx context.Context, repo repository.IRefreshTokenRepo, familyID uuid.UUID, now time.Time) error {
	return repo.RevokeByFamilyID(ctx, familyID, now)
}
