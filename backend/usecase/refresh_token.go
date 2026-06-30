package usecase

import (
	"context"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type RefreshTokenInput struct {
	RefreshToken string // Cookie から取得した平文
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string // ローテーション後の新平文
}

// RefreshToken は refresh を検証し、トークンローテーションで accessToken を再発行する
func RefreshToken(ctx context.Context, d AuthDeps, in RefreshTokenInput) (*RefreshTokenOutput, error) {
	if in.RefreshToken == "" {
		return nil, model.ErrUnauthorized()
	}
	now := d.now()
	hash := d.RefreshTokens.Hash(in.RefreshToken)

	stored, err := d.Repo.RefreshTokens().FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, model.ErrInternal("failed to find refresh token")
	}
	if stored == nil {
		return nil, model.ErrUnauthorized()
	}
	if stored.Status == model.RefreshTokenRevoked {
		if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
			return revokeRefreshTokenFamily(ctx, tx, stored.FamilyID, now)
		}); err != nil {
			return nil, err
		}
		if d.WSSessions != nil {
			_ = d.WSSessions.DeleteUser(ctx, stored.UserID)
		}
		return nil, model.ErrUnauthorized()
	}
	if !stored.IsActive(now) {
		if err := d.Repo.RefreshTokens().Revoke(ctx, stored.ID, now); err != nil {
			return nil, model.ErrInternal("failed to revoke expired refresh token")
		}
		return nil, model.ErrUnauthorized()
	}

	var accessToken, refreshPlain string
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		locked, err := tx.RefreshTokens().FindByTokenHashWithUserForUpdate(ctx, hash)
		if err != nil {
			return model.ErrInternal("failed to lock refresh token")
		}
		if locked == nil || !locked.IsActive(now) {
			return model.ErrUnauthorized()
		}
		user, err := tx.Users().FindByID(ctx, locked.UserID)
		if err != nil {
			return model.ErrInternal("failed to find user")
		}
		if user == nil || user.IsDeleted() {
			_ = tx.RefreshTokens().Revoke(ctx, locked.ID, now)
			return model.ErrUnauthorized()
		}

		accessToken, err = d.AccessTokens.Issue(user.ID, user.Role, now)
		if err != nil {
			return model.ErrInternal("failed to issue access token")
		}
		refreshPair, err := d.RefreshTokens.Generate()
		if err != nil {
			return model.ErrInternal("failed to generate refresh token")
		}
		newToken := model.NewRefreshToken(user.ID, refreshPair.Hash, locked.FamilyID, now.AddDate(0, 0, d.RefreshTokenExpiryDays), now)
		if err := tx.RefreshTokens().Create(ctx, &newToken); err != nil {
			return model.ErrInternal("failed to store refresh token")
		}
		if err := tx.RefreshTokens().MarkUsed(ctx, locked.ID, now, newToken.ID); err != nil {
			return model.ErrInternal("failed to rotate refresh token")
		}
		user.LastActivityAt = now
		user.UpdatedAt = now
		if err := tx.Users().Update(ctx, user); err != nil {
			return model.ErrInternal("failed to update user activity")
		}
		refreshPlain = refreshPair.Plaintext
		return nil
	}); err != nil {
		return nil, err
	}
	return &RefreshTokenOutput{AccessToken: accessToken, RefreshToken: refreshPlain}, nil
}
