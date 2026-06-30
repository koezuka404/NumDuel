package usecase

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type RefreshInput struct {
	RefreshToken string
}

func (a *AuthUseCase) Refresh(ctx context.Context, in RefreshInput) (*RefreshResult, error) {
	plain := strings.TrimSpace(in.RefreshToken)
	if plain == "" {
		return nil, ErrUnauthorized
	}
	now := a.now()
	hash := a.RefreshGen.Hash(plain)

	stored, err := a.RefreshTokens.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, ErrUnauthorized
	}
	if stored.Status == model.RefreshTokenRevoked {
		if err := a.handleReusedRefreshToken(ctx, stored, now); err != nil {
			return nil, err
		}
		return nil, ErrUnauthorized
	}
	if !stored.IsActive(now) {
		if err := a.revokeExpiredRefreshToken(ctx, stored.ID, now); err != nil {
			return nil, err
		}
		return nil, ErrUnauthorized
	}

	var result RefreshResult
	if err := repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		access, refreshPlain, err := a.rotateRefreshToken(ctx, hash, now)
		if err != nil {
			return err
		}
		result.AccessToken = access
		result.RefreshToken = refreshPlain
		return nil
	}); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AuthUseCase) handleReusedRefreshToken(ctx context.Context, stored *model.RefreshToken, now time.Time) error {
	err := repository.WithTx(ctx, a.DB, func(ctx context.Context) error {
		return revokeRefreshTokenFamily(ctx, a.RefreshTokens, stored.FamilyID, now)
	})
	if err != nil {
		return err
	}
	if a.WSSessions != nil {
		_ = a.WSSessions.DeleteUser(ctx, stored.UserID)
	}
	return nil
}

func (a *AuthUseCase) revokeExpiredRefreshToken(ctx context.Context, id uuid.UUID, now time.Time) error {
	return a.RefreshTokens.Revoke(ctx, id, now)
}

func (a *AuthUseCase) rotateRefreshToken(ctx context.Context, hash string, now time.Time) (string, string, error) {
	locked, err := a.RefreshTokens.FindByTokenHashForUpdate(ctx, hash)
	if err != nil {
		return "", "", err
	}
	if locked == nil || !locked.IsActive(now) {
		return "", "", ErrUnauthorized
	}
	user, err := a.Users.FindByID(ctx, locked.UserID)
	if err != nil {
		return "", "", err
	}
	if user == nil || user.IsDeleted() {
		_ = a.RefreshTokens.Revoke(ctx, locked.ID, now)
		return "", "", ErrUnauthorized
	}

	accessToken, err := a.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return "", "", err
	}
	pair, err := a.RefreshGen.Generate()
	if err != nil {
		return "", "", err
	}
	newToken := model.NewRefreshToken(user.ID, pair.Hash, locked.FamilyID, tokenExpiry(a, now), now)
	if err := a.RefreshTokens.Create(ctx, &newToken); err != nil {
		return "", "", err
	}
	if err := a.RefreshTokens.MarkUsed(ctx, locked.ID, now, newToken.ID); err != nil {
		return "", "", err
	}
	if err := a.touchUserActivity(ctx, user, now); err != nil {
		return "", "", err
	}
	return accessToken, pair.Plaintext, nil
}

func (a *AuthUseCase) CleanupExpiredRefreshTokens(ctx context.Context) {
	cutoff := a.now().AddDate(0, 0, -a.cleanupGraceDays())
	n, err := a.RefreshTokens.DeleteExpired(ctx, cutoff)
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("refresh token cleanup: %v", err)
		}
		return
	}
	if n > 0 {
		log.Printf("refresh token cleanup: deleted %d rows", n)
	}
}
