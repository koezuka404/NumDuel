package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
)

type IUserRepo interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ListAll(ctx context.Context) ([]*model.User, error)
	List(ctx context.Context, page, limit int) ([]*model.User, int64, error)
	Search(ctx context.Context, query string, page, limit int) ([]*model.User, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.User, error)
	ListInactiveSince(ctx context.Context, before time.Time) ([]*model.User, error)
	TouchLastActivity(ctx context.Context, userID uuid.UUID, at time.Time) error
	ExistsActiveMaster(ctx context.Context) (bool, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) IUserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) dbCtx(ctx context.Context) *gorm.DB {
	return dbFromCtx(ctx, r.db)
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	return r.dbCtx(ctx).Create(user).Error
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	return r.dbCtx(ctx).Save(user).Error
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return findOptional[model.User](r.dbCtx(ctx).Where("id = ?", id))
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return findOptional[model.User](r.dbCtx(ctx).Where("email = ?", email))
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return findOptional[model.User](r.dbCtx(ctx).Where("username = ?", username))
}

func (r *userRepo) ListAll(ctx context.Context) ([]*model.User, error) {
	var rows []model.User
	if err := r.dbCtx(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *userRepo) List(ctx context.Context, page, limit int) ([]*model.User, int64, error) {
	limit, offset := paginatePage(page, limit)
	var total int64
	q := r.dbCtx(ctx).Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
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

func (r *userRepo) Search(ctx context.Context, query string, page, limit int) ([]*model.User, int64, error) {
	pattern := "%" + query + "%"
	limit, offset := paginatePage(page, limit)
	var total int64
	q := userSearchScope(r.dbCtx(ctx), pattern)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.User
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

var isPostgresDialect = func(name string) bool { return name == "postgres" }

func userSearchScope(db *gorm.DB, pattern string) *gorm.DB {
	q := db.Model(&model.User{})
	if isPostgresDialect(db.Dialector.Name()) {
		return q.Where("username ILIKE ? OR email ILIKE ?", pattern, pattern)
	}
	return q.Where("lower(username) LIKE lower(?) OR lower(email) LIKE lower(?)", pattern, pattern)
}

func (r *userRepo) FindUpdatedSince(ctx context.Context, since time.Time) ([]*model.User, error) {
	var rows []model.User
	if err := r.dbCtx(ctx).Where("updated_at > ?", since).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *userRepo) ListInactiveSince(ctx context.Context, before time.Time) ([]*model.User, error) {
	var rows []model.User
	if err := r.dbCtx(ctx).
		Where("deleted_at IS NULL AND last_activity_at < ?", before).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.User, len(rows))
	for i := range rows {
		row := rows[i]
		out[i] = &row
	}
	return out, nil
}

func (r *userRepo) TouchLastActivity(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return r.dbCtx(ctx).Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"last_activity_at": at,
			"updated_at":       at,
		}).Error
}

func (r *userRepo) ExistsActiveMaster(ctx context.Context) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&model.User{}).
		Where("role = ? AND deleted_at IS NULL", model.RoleMaster).
		Count(&count).Error
	return count > 0, err
}
