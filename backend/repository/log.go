package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IActivityLogRepo interface {
	Create(ctx context.Context, log *model.ActivityLog) error
	Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]model.ActivityLog, int64, error)
	ListDistinctLogTypes(ctx context.Context) ([]string, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.ActivityLog, error)
}

type activityLogRepo struct {
	db *gorm.DB
}

func NewActivityLogRepo(db *gorm.DB) IActivityLogRepo {
	return &activityLogRepo{db: db}
}

func (r *activityLogRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *activityLogRepo) Create(ctx context.Context, log *model.ActivityLog) error {
	return r.dbCtx(ctx).Create(log).Error
}

func (r *activityLogRepo) Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]model.ActivityLog, int64, error) {
	q := r.dbCtx(ctx).Model(&model.ActivityLog{})
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
	limit, offset := paginatePage(page, limit)
	var rows []model.ActivityLog
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *activityLogRepo) ListDistinctLogTypes(ctx context.Context) ([]string, error) {
	var types []string
	err := r.dbCtx(ctx).Model(&model.ActivityLog{}).
		Distinct("log_type").Order("log_type ASC").Pluck("log_type", &types).Error
	return types, err
}

func (r *activityLogRepo) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.dbCtx(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&model.ActivityLog{})
	return res.RowsAffected, res.Error
}

func (r *activityLogRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.ActivityLog, error) {
	var rows []model.ActivityLog
	err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error
	return rows, err
}

type ILoginLogRepo interface {
	Create(ctx context.Context, log *model.LoginLog) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.LoginLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.LoginLog, error)
}

type loginLogRepo struct {
	db *gorm.DB
}

func NewLoginLogRepo(db *gorm.DB) ILoginLogRepo {
	return &loginLogRepo{db: db}
}

func (r *loginLogRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *loginLogRepo) Create(ctx context.Context, log *model.LoginLog) error {
	return r.dbCtx(ctx).Create(log).Error
}

func (r *loginLogRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.LoginLog, int64, error) {
	limit, offset := paginatePage(page, limit)
	var total int64
	q := r.dbCtx(ctx).Model(&model.LoginLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.LoginLog
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *loginLogRepo) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.dbCtx(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&model.LoginLog{})
	return res.RowsAffected, res.Error
}

func (r *loginLogRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.LoginLog, error) {
	var rows []model.LoginLog
	err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error
	return rows, err
}

type IWSConnectionLogRepo interface {
	Create(ctx context.Context, log *model.WSConnectionLog) error
	UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.WSConnectionLog, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error)
}

type wsConnectionLogRepo struct {
	db *gorm.DB
}

func NewWSConnectionLogRepo(db *gorm.DB) IWSConnectionLogRepo {
	return &wsConnectionLogRepo{db: db}
}

func (r *wsConnectionLogRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *wsConnectionLogRepo) Create(ctx context.Context, log *model.WSConnectionLog) error {
	return r.dbCtx(ctx).Create(log).Error
}

func (r *wsConnectionLogRepo) UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error {
	return r.dbCtx(ctx).Model(&model.WSConnectionLog{}).
		Where("id = ?", id).
		Update("disconnected_at", disconnectedAt).Error
}

func (r *wsConnectionLogRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.WSConnectionLog, int64, error) {
	limit, offset := paginatePage(page, limit)
	var total int64
	q := r.dbCtx(ctx).Model(&model.WSConnectionLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.WSConnectionLog
	if err := q.Order("connected_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *wsConnectionLogRepo) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.dbCtx(ctx).
		Where("connected_at < ?", before).
		Limit(batchSize).
		Delete(&model.WSConnectionLog{})
	return res.RowsAffected, res.Error
}
