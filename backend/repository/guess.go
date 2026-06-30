package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IGuessRepo interface {
	Create(ctx context.Context, guess *model.Guess) error
	ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]model.Guess, error)
	CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Guess, error)
}

type guessRepo struct {
	db *gorm.DB
}

func NewGuessRepo(db *gorm.DB) IGuessRepo {
	return &guessRepo{db: db}
}

func (r *guessRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *guessRepo) Create(ctx context.Context, guess *model.Guess) error {
	return r.dbCtx(ctx).Create(guess).Error
}

func (r *guessRepo) ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]model.Guess, error) {
	var rows []model.Guess
	err := r.dbCtx(ctx).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Order("turn ASC").
		Find(&rows).Error
	return rows, err
}

func (r *guessRepo) CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&model.Guess{}).
		Where("game_id = ? AND player_id <> ?", gameID, playerID).
		Count(&count).Error
	return count, err
}

func (r *guessRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Guess, error) {
	var rows []model.Guess
	err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error
	return rows, err
}
