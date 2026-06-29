package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type guessRepository struct{ db *gorm.DB }

func (r *guessRepository) Create(ctx context.Context, tx model.Transaction, guess *model.Guess) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(guess).Error
}

func (r *guessRepository) ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]model.Guess, error) {
	var rows []model.Guess
	err := r.db.WithContext(ctx).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Order("turn ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.Guess, 0, len(rows))
	out = append(out, rows...)
	return out, nil
}

func (r *guessRepository) CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Guess{}).
		Where("game_id = ? AND player_id <> ?", gameID, playerID).
		Count(&count).Error
	return count, err
}

func (r *guessRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Guess, error) {
	var rows []model.Guess
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.Guess, 0, len(rows))
	out = append(out, rows...)
	return out, nil
}
