package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type activityLogRepository struct{ db *gorm.DB }

func (r *activityLogRepository) Create(ctx context.Context, log *model.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *activityLogRepository) Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]model.ActivityLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.ActivityLog{})
	if logType != "" {
		q = q.Where("log_type = ?", logType)
	}
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ActivityLog
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.ActivityLog, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, total, nil
}

func (r *activityLogRepository) ListDistinctLogTypes(ctx context.Context) ([]string, error) {
	var types []string
	err := r.db.WithContext(ctx).Model(&model.ActivityLog{}).
		Distinct("log_type").Order("log_type ASC").Pluck("log_type", &types).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, len(types))
	copy(out, types)
	return out, nil
}

func (r *activityLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&model.ActivityLog{})
	return res.RowsAffected, res.Error
}

func (r *activityLogRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.ActivityLog, error) {
	var rows []model.ActivityLog
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.ActivityLog, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}
