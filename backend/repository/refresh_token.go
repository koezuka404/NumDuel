package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IRefreshTokenRepo interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	FindByTokenHashForUpdate(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error
	Create(ctx context.Context, token *model.RefreshToken) error
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepo(db *gorm.DB) IRefreshTokenRepo {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *refreshTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	return findOptional[model.RefreshToken](r.dbCtx(ctx).Where("token_hash = ?", tokenHash))
}

func (r *refreshTokenRepo) FindByTokenHashForUpdate(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	return findOptionalForUpdate[model.RefreshToken](r.dbCtx(ctx).Where("token_hash = ?", tokenHash))
}

func (r *refreshTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time, replacedByTokenID uuid.UUID) error {
	res := r.dbCtx(ctx).Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":               model.RefreshTokenRevoked,
			"revoked_at":           usedAt,
			"replaced_by_token_id": replacedByTokenID,
			"updated_at":           usedAt,
		})
	return rowsAffected(res.Error, res.RowsAffected)
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.dbCtx(ctx).Create(token).Error
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	res := r.dbCtx(ctx).Model(&model.RefreshToken{}).
		Where("id = ? AND status = ? AND revoked_at IS NULL", id, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		})
	return rowsAffected(res.Error, res.RowsAffected)
}

func (r *refreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error {
	return r.dbCtx(ctx).Model(&model.RefreshToken{}).
		Where("family_id = ? AND status = ? AND revoked_at IS NULL", familyID, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		}).Error
}

func (r *refreshTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.dbCtx(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ? AND status = ? AND revoked_at IS NULL", userID, model.RefreshTokenActive).
		Updates(map[string]any{
			"status":     model.RefreshTokenRevoked,
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		}).Error
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res := r.dbCtx(ctx).
		Where("(status = ? AND expires_at < ?) OR (status = ? AND revoked_at IS NOT NULL AND revoked_at < ?)",
			model.RefreshTokenActive, now,
			model.RefreshTokenRevoked, now).
		Delete(&model.RefreshToken{})
	return res.RowsAffected, res.Error
}
