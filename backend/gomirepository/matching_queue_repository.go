package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type matchingQueueRepository struct{ db *gorm.DB }

func (r *matchingQueueRepository) Insert(ctx context.Context, entry *model.MatchingQueueEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *matchingQueueRepository) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&model.MatchingQueueEntry{}, "id IN ?", ids).Error
}

func (r *matchingQueueRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MatchingQueueEntry{}, "user_id = ?", userID).Error
}

func (r *matchingQueueRepository) ListByStatusForUpdate(ctx context.Context, status model.MatchingQueueStatus, limit int) ([]model.MatchingQueueEntry, error) {
	var rows []model.MatchingQueueEntry
	err := forUpdate(r.db.WithContext(ctx)).
		Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.MatchingQueueEntry, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}

func (r *matchingQueueRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.MatchingQueueEntry, error) {
	var m model.MatchingQueueEntry
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
