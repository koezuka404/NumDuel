package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

// IRefreshTokenRepository は refresh_tokens テーブルへのアクセス
type GormRefreshTokenRepository struct {
	baseRepo
}

var _ IRefreshTokenRepository = (*GormRefreshTokenRepository)(nil)

// NewRefreshTokenRepository は RefreshTokenRepository を作成する
func NewRefreshTokenRepository(db *gorm.DB) IRefreshTokenRepository {
	return &GormRefreshTokenRepository{baseRepo{db: db}}
}

func (r *GormRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		if err := mapDBError(err); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &token, nil
}

func (r *GormRefreshTokenRepository) FindByTokenHashWithUser(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ?", tokenHash).
		First(&token).Error; err != nil {
		if err := mapDBError(err); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &token, nil
}

func (r *GormRefreshTokenRepository) FindByTokenHashWithUserForUpdate(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := forUpdate(r.db.WithContext(ctx)).
		Preload("User").
		Where("token_hash = ?", tokenHash).
		First(&token).Error; err != nil {
		if err := mapDBError(err); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &token, nil
}

func (r *GormRefreshTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":               model.RefreshTokenRevoked,
			"revoked_at":           usedAt,
			"replaced_by_token_id": replacedByTokenID,
			"updated_at":           usedAt,
		})
	return mapRowsAffected(res.RowsAffected, res.Error)
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Omit("User").Create(token).Error
}

func (r *GormRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return mapRowsAffected(res.RowsAffected, res.Error)
}

func (r *GormRefreshTokenRepository) RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("family_id = ? AND status = ? AND revoked_at IS NULL", familyID, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return res.Error
}

func (r *GormRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ? AND status = ? AND revoked_at IS NULL", userID, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return res.Error
}

func (r *GormRefreshTokenRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("(status = ? AND expires_at < ?) OR (status = ? AND revoked_at IS NOT NULL AND revoked_at < ?)",
			model.RefreshTokenActive, now,
			model.RefreshTokenRevoked, now).
		Delete(&model.RefreshToken{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
