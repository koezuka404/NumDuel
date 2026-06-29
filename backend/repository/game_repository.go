package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type gameRepository struct{ db *gorm.DB }

func (r *gameRepository) Create(ctx context.Context, game *model.Game) error {
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *gameRepository) Update(ctx context.Context, game *model.Game) error {
	return r.db.WithContext(ctx).Save(game).Error
}

func (r *gameRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	var m model.Game
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *gameRepository) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	var m model.Game
	err := forUpdate(r.db.WithContext(ctx)).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *gameRepository) ListByPlayerID(ctx context.Context, userID uuid.UUID) ([]*model.Game, error) {
	var rows []model.Game
	err := r.db.WithContext(ctx).
		Where("player1_id = ? OR player2_id = ?", userID, userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.Game, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *gameRepository) ListByStatus(ctx context.Context, status model.GameStatus) ([]*model.Game, error) {
	var rows []model.Game
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.Game, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *gameRepository) ListByStatusCreatedBefore(ctx context.Context, status model.GameStatus, before time.Time) ([]*model.Game, error) {
	var rows []model.Game
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", status, before).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.Game, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *gameRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.Game, error) {
	var rows []model.Game
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.Game, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}
