package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type loginLogRepository struct{ db *gorm.DB }

func (r *loginLogRepository) Create(ctx context.Context, log *model.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *loginLogRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.LoginLog, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.LoginLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.LoginLog
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.LoginLog, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, total, nil
}

func (r *loginLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&model.LoginLog{})
	return res.RowsAffected, res.Error
}

func (r *loginLogRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.LoginLog, error) {
	var rows []model.LoginLog
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.LoginLog, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}
