package usecase

import (
	"context"

	"github.com/numduel/numduel/model"
)

type RefreshTokenInput struct {
	RefreshToken string // Cookie から取得した平文
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string // ローテーション後の新平文
}

// RefreshToken は refresh を検証し、トークンローテーションで accessToken を再発行する。
func RefreshToken(ctx context.Context, d AuthDeps, in RefreshTokenInput) (*RefreshTokenOutput, error) {
	if in.RefreshToken == "" {
		return nil, model.ErrUnauthorized()
	}
	now := d.now()
	stored, err := d.Repo.RefreshTokens().FindByTokenHash(ctx, d.RefreshTokens.Hash(in.RefreshToken))
	if err != nil {
		return nil, model.ErrInternal("failed to find refresh token")
	}
	if stored == nil {
		return nil, model.ErrUnauthorized()
	}
	// 失効済みトークンの再使用 = 盗用疑い → family 一括失効
	if stored.Status == model.RefreshTokenRevoked {
		if err := withTx(ctx, d.Repo, func(tx model.Transaction) error {
			return revokeRefreshTokenFamily(ctx, d.Repo, tx, stored.FamilyID, now)
		}); err != nil {
			return nil, err
		}
		if d.WSSessions != nil {
			_ = d.WSSessions.DeleteUser(ctx, stored.UserID)
		}
		return nil, model.ErrUnauthorized()
	}
	if !stored.IsActive(now) {
		stored.Revoke(now)
		if err := d.Repo.RefreshTokens().Update(ctx, nil, stored); err != nil {
			return nil, model.ErrInternal("failed to revoke expired refresh token")
		}
		return nil, model.ErrUnauthorized()
	}
	user, err := d.Repo.Users().FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, model.ErrInternal("failed to find user")
	}
	if user == nil || user.IsDeleted() {
		stored.Revoke(now)
		_ = d.Repo.RefreshTokens().Update(ctx, nil, stored)
		return nil, model.ErrUnauthorized()
	}
	accessToken, err := d.AccessTokens.Issue(user.ID, user.Role, now)
	if err != nil {
		return nil, model.ErrInternal("failed to issue access token")
	}
	refreshPair, err := d.RefreshTokens.Generate()
	if err != nil {
		return nil, model.ErrInternal("failed to generate refresh token")
	}
	newToken := model.NewRefreshToken(user.ID, refreshPair.Hash, stored.FamilyID, now.AddDate(0, 0, d.RefreshTokenExpiryDays), now)
	// 旧 refresh を失効 → 新 refresh を INSERT（ローテーション）
	if err := withTx(ctx, d.Repo, func(tx model.Transaction) error {
		stored.Revoke(now)
		if err := d.Repo.RefreshTokens().Update(ctx, tx, stored); err != nil {
			return model.ErrInternal("failed to revoke old refresh token")
		}
		if err := d.Repo.RefreshTokens().Create(ctx, tx, &newToken); err != nil {
			return model.ErrInternal("failed to store refresh token")
		}
		user.LastActivityAt = now
		user.UpdatedAt = now
		if err := d.Repo.Users().Update(ctx, tx, user); err != nil {
			return model.ErrInternal("failed to update user activity")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &RefreshTokenOutput{AccessToken: accessToken, RefreshToken: refreshPair.Plaintext}, nil
}
