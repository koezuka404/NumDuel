package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

// IRefreshTokenRepository は RefreshToken の検索・rotation・失効・削除。
type IRefreshTokenRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	FindByTokenHashWithUser(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	FindByTokenHashWithUserForUpdate(ctx context.Context, tx model.Transaction, tokenHash string) (*model.RefreshToken, error)
	MarkUsed(ctx context.Context, tx model.Transaction, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error
	Create(ctx context.Context, tx model.Transaction, token *model.RefreshToken) error
	Revoke(ctx context.Context, tx model.Transaction, id uuid.UUID, revokedAt time.Time) error
	RevokeByFamilyID(ctx context.Context, tx model.Transaction, familyID uuid.UUID, revokedAt time.Time) error
	RevokeByUserID(ctx context.Context, tx model.Transaction, userID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

type GormRefreshTokenRepository struct {
	baseRepo
}

var (
	_ IRefreshTokenRepository      = (*GormRefreshTokenRepository)(nil)
	_ model.RefreshTokenRepository = (*GormRefreshTokenRepository)(nil)
)

// NewRefreshTokenRepository は RefreshTokenRepository を作成する。
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

func (r *GormRefreshTokenRepository) FindByTokenHashWithUserForUpdate(ctx context.Context, tx model.Transaction, tokenHash string) (*model.RefreshToken, error) {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return nil, err
	}
	var token model.RefreshToken
	if err := forUpdate(db).
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

func (r *GormRefreshTokenRepository) MarkUsed(ctx context.Context, tx model.Transaction, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	res := db.Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":               model.RefreshTokenRevoked,
			"revoked_at":           usedAt,
			"replaced_by_token_id": replacedByTokenID,
			"updated_at":           usedAt,
		})
	return mapRowsAffected(res.RowsAffected, res.Error)
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, tx model.Transaction, token *model.RefreshToken) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Omit("User").Create(token).Error
}

func (r *GormRefreshTokenRepository) Revoke(ctx context.Context, tx model.Transaction, id uuid.UUID, revokedAt time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	res := db.Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return mapRowsAffected(res.RowsAffected, res.Error)
}

func (r *GormRefreshTokenRepository) RevokeByFamilyID(ctx context.Context, tx model.Transaction, familyID uuid.UUID, revokedAt time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	res := db.Model(&model.RefreshToken{}).
		Where("family_id = ? AND status = ? AND revoked_at IS NULL", familyID, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return res.Error
}

func (r *GormRefreshTokenRepository) RevokeByUserID(ctx context.Context, tx model.Transaction, userID uuid.UUID, revokedAt time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	res := db.Model(&model.RefreshToken{}).
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
