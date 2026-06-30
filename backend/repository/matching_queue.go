package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/numduel/numduel/model"
)

type IMatchingQueueRepo interface {
	Insert(ctx context.Context, entry *model.MatchingQueueEntry) error
	DeleteByIDs(ctx context.Context, ids []uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	ListByStatusForUpdate(ctx context.Context, status model.MatchingQueueStatus, limit int) ([]model.MatchingQueueEntry, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.MatchingQueueEntry, error)
}

type matchingQueueRepo struct {
	db *gorm.DB
}

func NewMatchingQueueRepo(db *gorm.DB) IMatchingQueueRepo {
	return &matchingQueueRepo{db: db}
}

func (r *matchingQueueRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *matchingQueueRepo) Insert(ctx context.Context, entry *model.MatchingQueueEntry) error {
	return r.dbCtx(ctx).Create(entry).Error
}

func (r *matchingQueueRepo) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.dbCtx(ctx).Delete(&model.MatchingQueueEntry{}, "id IN ?", ids).Error
}

func (r *matchingQueueRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.dbCtx(ctx).Delete(&model.MatchingQueueEntry{}, "user_id = ?", userID).Error
}

func (r *matchingQueueRepo) ListByStatusForUpdate(ctx context.Context, status model.MatchingQueueStatus, limit int) ([]model.MatchingQueueEntry, error) {
	var rows []model.MatchingQueueEntry
	err := r.dbCtx(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *matchingQueueRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.MatchingQueueEntry, error) {
	return findOptional[model.MatchingQueueEntry](
		r.dbCtx(ctx).Where("user_id = ?", userID).Order("created_at DESC"),
	)
}
