package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type matchHistoryRepository struct{ db *gorm.DB }

func (r *matchHistoryRepository) Create(ctx context.Context, tx model.Transaction, history *model.MatchHistory) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(history).Error
}

func (r *matchHistoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.MatchHistory, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.MatchHistory{}).
		Where("winner_id = ? OR loser_id = ?", userID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.MatchHistory
	offset := (page - 1) * limit
	if err := q.Order("finished_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.MatchHistory, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, total, nil
}

func (r *matchHistoryRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.MatchHistory, error) {
	var rows []model.MatchHistory
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.MatchHistory, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}
