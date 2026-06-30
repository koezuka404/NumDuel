package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IMatchHistoryRepo interface {
	Create(ctx context.Context, history *model.MatchHistory) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.MatchHistory, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.MatchHistory, error)
}

type matchHistoryRepo struct {
	db *gorm.DB
}

func NewMatchHistoryRepo(db *gorm.DB) IMatchHistoryRepo {
	return &matchHistoryRepo{db: db}
}

func (r *matchHistoryRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *matchHistoryRepo) Create(ctx context.Context, history *model.MatchHistory) error {
	return r.dbCtx(ctx).Create(history).Error
}

func (r *matchHistoryRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.MatchHistory, int64, error) {
	limit, offset := paginatePage(page, limit)
	var total int64
	q := r.dbCtx(ctx).Model(&model.MatchHistory{}).
		Where("winner_id = ? OR loser_id = ?", userID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.MatchHistory
	if err := q.Order("finished_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *matchHistoryRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.MatchHistory, error) {
	var rows []model.MatchHistory
	err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error
	return rows, err
}
