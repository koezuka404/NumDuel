package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type matchingQueueRepository struct{ db *gorm.DB }

func (r *matchingQueueRepository) Insert(ctx context.Context, tx model.Transaction, entry *model.MatchingQueueEntry) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(entry).Error
}

func (r *matchingQueueRepository) DeleteByIDs(ctx context.Context, tx model.Transaction, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.MatchingQueueEntry{}, "id IN ?", ids).Error
}

func (r *matchingQueueRepository) DeleteByUserID(ctx context.Context, tx model.Transaction, userID uuid.UUID) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Delete(&model.MatchingQueueEntry{}, "user_id = ?", userID).Error
}

func (r *matchingQueueRepository) ListByStatusForUpdate(ctx context.Context, tx model.Transaction, status model.MatchingQueueStatus, limit int) ([]model.MatchingQueueEntry, error) {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return nil, err
	}
	var rows []model.MatchingQueueEntry
	err = forUpdate(db).
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
