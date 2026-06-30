package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IGameRepo interface {
	Create(ctx context.Context, game *model.Game) error
	Update(ctx context.Context, game *model.Game) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error)
	FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListByPlayerID(ctx context.Context, userID uuid.UUID) ([]*model.Game, error)
	ListByStatus(ctx context.Context, status model.GameStatus) ([]*model.Game, error)
	ListByStatusCreatedBefore(ctx context.Context, status model.GameStatus, before time.Time) ([]*model.Game, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.Game, error)
}

type gameRepo struct {
	db *gorm.DB
}

func NewGameRepo(db *gorm.DB) IGameRepo {
	return &gameRepo{db: db}
}

func (r *gameRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *gameRepo) Create(ctx context.Context, game *model.Game) error {
	return r.dbCtx(ctx).Create(game).Error
}

func (r *gameRepo) Update(ctx context.Context, game *model.Game) error {
	return r.dbCtx(ctx).Save(game).Error
}

func (r *gameRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	return findOptional[model.Game](r.dbCtx(ctx).Where("id = ?", id))
}

func (r *gameRepo) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	return findOptionalForUpdate[model.Game](r.dbCtx(ctx).Where("id = ?", id))
}

func (r *gameRepo) ListByPlayerID(ctx context.Context, userID uuid.UUID) ([]*model.Game, error) {
	var rows []model.Game
	if err := r.dbCtx(ctx).
		Where("player1_id = ? OR player2_id = ?", userID, userID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return gameRowsToPtrs(rows), nil
}

func (r *gameRepo) ListByStatus(ctx context.Context, status model.GameStatus) ([]*model.Game, error) {
	var rows []model.Game
	if err := r.dbCtx(ctx).Where("status = ?", status).Find(&rows).Error; err != nil {
		return nil, err
	}
	return gameRowsToPtrs(rows), nil
}

func (r *gameRepo) ListByStatusCreatedBefore(ctx context.Context, status model.GameStatus, before time.Time) ([]*model.Game, error) {
	var rows []model.Game
	if err := r.dbCtx(ctx).
		Where("status = ? AND created_at < ?", status, before).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return gameRowsToPtrs(rows), nil
}

func (r *gameRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.Game, error) {
	var rows []model.Game
	if err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error; err != nil {
		return nil, err
	}
	return gameRowsToPtrs(rows), nil
}

func gameRowsToPtrs(rows []model.Game) []*model.Game {
	out := make([]*model.Game, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out
}
