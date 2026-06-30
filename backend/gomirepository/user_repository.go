package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type userRepository struct{ db *gorm.DB }

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var m model.User
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var m model.User
	err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var m model.User
	err := r.db.WithContext(ctx).First(&m, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	var rows []model.User
	err := r.db.WithContext(ctx).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *userRepository) List(ctx context.Context, page, limit int) ([]*model.User, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, total, nil
}

func (r *userRepository) Search(ctx context.Context, query string, page, limit int) ([]*model.User, int64, error) {
	pattern := "%" + query + "%"
	var total int64
	q := r.db.WithContext(ctx).Model(&model.User{}).
		Where("username ILIKE ? OR email ILIKE ?", pattern, pattern)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, total, nil
}

func (r *userRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.User, error) {
	var rows []model.User
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

// ListInactiveSince は last_activity_at が before より古い未削除ユーザーを返す（AutoLogoutWorker 用）
func (r *userRepository) ListInactiveSince(ctx context.Context, before time.Time) ([]*model.User, error) {
	var rows []model.User
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND last_activity_at < ?", before).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

// TouchLastActivity は last_activity_at / updated_at を更新する（ActivityUpdateMiddleware 用）
// FindByID せず UPDATE のみ。deleted_at IS NULL のユーザーのみ対象
func (r *userRepository) TouchLastActivity(ctx context.Context, userID uuid.UUID, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"last_activity_at": at,
			"updated_at":       at,
		})
	return res.Error
}

// ExistsActiveMaster は未削除の master ユーザーが存在するか
func (r *userRepository) ExistsActiveMaster(ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("role = ? AND deleted_at IS NULL", model.RoleMaster).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
