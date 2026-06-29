package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type refreshTokenRepository struct{ db *gorm.DB }

func (r *refreshTokenRepository) Create(ctx context.Context, tx model.Transaction, token *model.RefreshToken) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(token).Error
}

func (r *refreshTokenRepository) Update(ctx context.Context, tx model.Transaction, token *model.RefreshToken) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Save(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var m model.RefreshToken
	err := r.db.WithContext(ctx).First(&m, "token_hash = ?", tokenHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *refreshTokenRepository) UpdateStatusByUserID(ctx context.Context, tx model.Transaction, userID uuid.UUID, fromStatus, toStatus model.RefreshTokenStatus, revokedAt *time.Time, now time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	updates := map[string]any{"status": toStatus, "updated_at": now}
	if revokedAt != nil {
		updates["revoked_at"] = *revokedAt
	}
	return db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND status = ?", userID, fromStatus).
		Updates(updates).Error
}

func (r *refreshTokenRepository) UpdateStatusByFamilyID(ctx context.Context, tx model.Transaction, familyID uuid.UUID, fromStatus, toStatus model.RefreshTokenStatus, revokedAt *time.Time, now time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	updates := map[string]any{"status": toStatus, "updated_at": now}
	if revokedAt != nil {
		updates["revoked_at"] = *revokedAt
	}
	return db.Model(&model.RefreshToken{}).
		Where("family_id = ? AND status = ?", familyID, fromStatus).
		Updates(updates).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("(status = ? AND expires_at < ?) OR (status = ? AND revoked_at < ?)",
			model.RefreshTokenActive, before,
			model.RefreshTokenRevoked, before).
		Delete(&model.RefreshToken{})
	return res.RowsAffected, res.Error
}
