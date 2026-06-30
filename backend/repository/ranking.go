package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IRankingRepo interface {
	ReplaceAll(ctx context.Context, rankings []model.Ranking) error
	ListAll(ctx context.Context) ([]model.Ranking, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Ranking, error)
}

type rankingRepo struct {
	db *gorm.DB
}

func NewRankingRepo(db *gorm.DB) IRankingRepo {
	return &rankingRepo{db: db}
}

func (r *rankingRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *rankingRepo) ReplaceAll(ctx context.Context, rankings []model.Ranking) error {
	if err := r.dbCtx(ctx).Exec("DELETE FROM rankings").Error; err != nil {
		return err
	}
	if len(rankings) == 0 {
		return nil
	}
	return r.dbCtx(ctx).Create(&rankings).Error
}

func (r *rankingRepo) ListAll(ctx context.Context) ([]model.Ranking, error) {
	var rows []model.Ranking
	err := r.dbCtx(ctx).Order("rank ASC").Find(&rows).Error
	return rows, err
}

func (r *rankingRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Ranking, error) {
	var rows []model.Ranking
	err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error
	return rows, err
}
