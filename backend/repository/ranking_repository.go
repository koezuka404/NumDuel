package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type rankingRepository struct{ db *gorm.DB }

func (r *rankingRepository) ReplaceAll(ctx context.Context, tx model.Transaction, rankings []model.Ranking) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM rankings").Error; err != nil {
		return err
	}
	if len(rankings) == 0 {
		return nil
	}
	return db.Create(&rankings).Error
}

func (r *rankingRepository) ListAll(ctx context.Context) ([]model.Ranking, error) {
	var rows []model.Ranking
	err := r.db.WithContext(ctx).Order("rank ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.Ranking, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}

func (r *rankingRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Ranking, error) {
	var rows []model.Ranking
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.Ranking, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}
