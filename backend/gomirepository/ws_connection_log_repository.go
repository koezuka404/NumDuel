package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type wsConnectionLogRepository struct{ db *gorm.DB }

func (r *wsConnectionLogRepository) Create(ctx context.Context, log *model.WSConnectionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *wsConnectionLogRepository) UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.WSConnectionLog{}).
		Where("id = ?", id).
		Update("disconnected_at", disconnectedAt).Error
}

func (r *wsConnectionLogRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.WSConnectionLog, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.WSConnectionLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.WSConnectionLog
	offset := (page - 1) * limit
	if err := q.Order("connected_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.WSConnectionLog, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, total, nil
}

func (r *wsConnectionLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("connected_at < ?", before).
		Limit(batchSize).
		Delete(&model.WSConnectionLog{})
	return res.RowsAffected, res.Error
}
