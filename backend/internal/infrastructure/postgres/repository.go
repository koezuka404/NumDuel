package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/numduel/numduel/internal/domain"
)

type Repository struct {
	db *DB

	users            *userRepository
	games            *gameRepository
	guesses          *guessRepository
	matchHistories   *matchHistoryRepository
	matchingQueue    *matchingQueueRepository
	rankings         *rankingRepository
	refreshTokens    *refreshTokenRepository
	activityLogs     *activityLogRepository
	loginLogs        *loginLogRepository
	wsConnectionLogs *wsConnectionLogRepository
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *DB) *Repository {
	g := db.Gorm()
	return &Repository{
		db:               db,
		users:            &userRepository{db: g},
		games:            &gameRepository{db: g},
		guesses:          &guessRepository{db: g},
		matchHistories:   &matchHistoryRepository{db: g},
		matchingQueue:    &matchingQueueRepository{db: g},
		rankings:         &rankingRepository{db: g},
		refreshTokens:    &refreshTokenRepository{db: g},
		activityLogs:     &activityLogRepository{db: g},
		loginLogs:        &loginLogRepository{db: g},
		wsConnectionLogs: &wsConnectionLogRepository{db: g},
	}
}

func (r *Repository) Begin(ctx context.Context) (domain.Transaction, error) {
	return r.db.Begin(ctx)
}

func (r *Repository) Commit(tx domain.Transaction) error {
	return r.db.Commit(tx)
}

func (r *Repository) Rollback(tx domain.Transaction) error {
	return r.db.Rollback(tx)
}

func (r *Repository) Users() domain.UserRepository                 { return r.users }
func (r *Repository) Games() domain.GameRepository                 { return r.games }
func (r *Repository) Guesses() domain.GuessRepository              { return r.guesses }
func (r *Repository) MatchHistories() domain.MatchHistoryRepository { return r.matchHistories }
func (r *Repository) MatchingQueue() domain.MatchingQueueRepository { return r.matchingQueue }
func (r *Repository) Rankings() domain.RankingRepository           { return r.rankings }
func (r *Repository) RefreshTokens() domain.RefreshTokenRepository { return r.refreshTokens }
func (r *Repository) ActivityLogs() domain.ActivityLogRepository   { return r.activityLogs }
func (r *Repository) LoginLogs() domain.LoginLogRepository         { return r.loginLogs }
func (r *Repository) WSConnectionLogs() domain.WSConnectionLogRepository {
	return r.wsConnectionLogs
}

func conn(ctx context.Context, db *gorm.DB, tx domain.Transaction) (*gorm.DB, error) {
	base, err := dbOrGlobal(db.WithContext(ctx), tx)
	if err != nil {
		return nil, err
	}
	return base, nil
}

type userRepository struct{ db *gorm.DB }

func (r *userRepository) Create(ctx context.Context, tx domain.Transaction, user *domain.User) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromUser(user)).Error
}

func (r *userRepository) Update(ctx context.Context, tx domain.Transaction, user *domain.User) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Save(fromUser(user)).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&m), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&m), nil
}

func (r *userRepository) FindByEmailActive(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).First(&m, "email = ? AND deleted_at IS NULL", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&m), nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).First(&m, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&m), nil
}

func (r *userRepository) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("email = ? OR username = ?", email, username).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) CountMasters(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("role = ?", domain.RoleMaster).
		Count(&count).Error
	return count, err
}

func (r *userRepository) ListForRankingRebuild(ctx context.Context) ([]domain.RankingRebuildRow, error) {
	var rows []userModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND role <> ?", domain.RoleMaster).
		Order("win_count DESC, username ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.RankingRebuildRow, len(rows))
	for i := range rows {
		out[i] = domain.RankingRebuildRow{
			UserID:   rows[i].ID,
			Username: rows[i].Username,
			WinCount: rows[i].WinCount,
		}
	}
	return out, nil
}

func (r *userRepository) IncrementWinCount(ctx context.Context, tx domain.Transaction, userID uuid.UUID, now time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Model(&userModel{}).Where("id = ?", userID).
		Updates(map[string]any{
			"win_count":  gorm.Expr("win_count + 1"),
			"updated_at": now,
		}).Error
}

func (r *userRepository) FindInactive(ctx context.Context, inactiveBefore time.Time) ([]*domain.User, error) {
	var rows []userModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND last_activity_at < ?", inactiveBefore).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = toUser(&rows[i])
	}
	return out, nil
}

func (r *userRepository) List(ctx context.Context, page, limit int) ([]*domain.User, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&userModel{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []userModel
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = toUser(&rows[i])
	}
	return out, total, nil
}

func (r *userRepository) Search(ctx context.Context, query string, page, limit int) ([]*domain.User, int64, error) {
	pattern := "%" + query + "%"
	var total int64
	q := r.db.WithContext(ctx).Model(&userModel{}).
		Where("username ILIKE ? OR email ILIKE ?", pattern, pattern)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []userModel
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = toUser(&rows[i])
	}
	return out, total, nil
}

func (r *userRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]*domain.User, error) {
	var rows []userModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = toUser(&rows[i])
	}
	return out, nil
}

type gameRepository struct{ db *gorm.DB }

func (r *gameRepository) Create(ctx context.Context, tx domain.Transaction, game *domain.Game) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromGame(game)).Error
}

func (r *gameRepository) Update(ctx context.Context, tx domain.Transaction, game *domain.Game) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Save(fromGame(game)).Error
}

func (r *gameRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	var m gameModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toGame(&m), nil
}

func (r *gameRepository) FindByIDForUpdate(ctx context.Context, tx domain.Transaction, id uuid.UUID) (*domain.Game, error) {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return nil, err
	}
	var m gameModel
	err = forUpdate(db).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toGame(&m), nil
}

func (r *gameRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.Game, error) {
	var m gameModel
	err := r.db.WithContext(ctx).
		Where("(player1_id = ? OR player2_id = ?) AND status IN ?", userID, userID,
			[]string{string(domain.GameStatusWaitingSecret), string(domain.GameStatusInProgress)}).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toGame(&m), nil
}

func (r *gameRepository) FindAllInProgress(ctx context.Context) ([]*domain.Game, error) {
	var rows []gameModel
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.GameStatusInProgress).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Game, len(rows))
	for i := range rows {
		out[i] = toGame(&rows[i])
	}
	return out, nil
}

func (r *gameRepository) FindWaitingSecretExpired(ctx context.Context, deadline time.Time) ([]*domain.Game, error) {
	var rows []gameModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", domain.GameStatusWaitingSecret, deadline).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Game, len(rows))
	for i := range rows {
		out[i] = toGame(&rows[i])
	}
	return out, nil
}

func (r *gameRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]*domain.Game, error) {
	var rows []gameModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Game, len(rows))
	for i := range rows {
		out[i] = toGame(&rows[i])
	}
	return out, nil
}

type guessRepository struct{ db *gorm.DB }

func (r *guessRepository) Create(ctx context.Context, tx domain.Transaction, guess *domain.Guess) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	m, err := fromGuess(guess)
	if err != nil {
		return err
	}
	return db.Create(m).Error
}

func (r *guessRepository) ListByGameAndPlayer(ctx context.Context, gameID, playerID uuid.UUID) ([]domain.Guess, error) {
	var rows []guessModel
	err := r.db.WithContext(ctx).
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Order("turn ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Guess, 0, len(rows))
	for i := range rows {
		g, err := toGuess(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *guessRepository) CountByGameExcludingPlayer(ctx context.Context, gameID, playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&guessModel{}).
		Where("game_id = ? AND player_id <> ?", gameID, playerID).
		Count(&count).Error
	return count, err
}

func (r *guessRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]domain.Guess, error) {
	var rows []guessModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Guess, 0, len(rows))
	for i := range rows {
		g, err := toGuess(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

type matchHistoryRepository struct{ db *gorm.DB }

func (r *matchHistoryRepository) Create(ctx context.Context, tx domain.Transaction, history *domain.MatchHistory) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromMatchHistory(history)).Error
}

func (r *matchHistoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.MatchHistory, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&matchHistoryModel{}).
		Where("winner_id = ? OR loser_id = ?", userID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []matchHistoryModel
	offset := (page - 1) * limit
	if err := q.Order("finished_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.MatchHistory, len(rows))
	for i := range rows {
		out[i] = toMatchHistory(&rows[i])
	}
	return out, total, nil
}

func (r *matchHistoryRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]domain.MatchHistory, error) {
	var rows []matchHistoryModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.MatchHistory, len(rows))
	for i := range rows {
		out[i] = toMatchHistory(&rows[i])
	}
	return out, nil
}

type matchingQueueRepository struct{ db *gorm.DB }

func (r *matchingQueueRepository) Insert(ctx context.Context, tx domain.Transaction, entry *domain.MatchingQueueEntry) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromMatchingQueueEntry(entry)).Error
}

func (r *matchingQueueRepository) DeleteByIDs(ctx context.Context, tx domain.Transaction, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Delete(&matchingQueueModel{}, "id IN ?", ids).Error
}

func (r *matchingQueueRepository) DeleteByUserID(ctx context.Context, tx domain.Transaction, userID uuid.UUID) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Delete(&matchingQueueModel{}, "user_id = ?", userID).Error
}

func (r *matchingQueueRepository) FindWaitingForUpdate(ctx context.Context, tx domain.Transaction, limit int) ([]domain.MatchingQueueEntry, error) {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return nil, err
	}
	var rows []matchingQueueModel
	err = forUpdate(db).
		Where("status = ?", domain.MatchingQueueWaiting).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.MatchingQueueEntry, len(rows))
	for i := range rows {
		out[i] = toMatchingQueueEntry(&rows[i])
	}
	return out, nil
}

func (r *matchingQueueRepository) ExistsWaitingByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&matchingQueueModel{}).
		Where("user_id = ? AND status = ?", userID, domain.MatchingQueueWaiting).
		Count(&count).Error
	return count > 0, err
}

func (r *matchingQueueRepository) FindWaitingByUserID(ctx context.Context, userID uuid.UUID) (*domain.MatchingQueueEntry, error) {
	var m matchingQueueModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.MatchingQueueWaiting).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entry := toMatchingQueueEntry(&m)
	return &entry, nil
}

type rankingRepository struct{ db *gorm.DB }

func (r *rankingRepository) ReplaceAll(ctx context.Context, tx domain.Transaction, rankings []domain.Ranking) error {
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
	rows := make([]rankingModel, len(rankings))
	for i, item := range rankings {
		rows[i] = fromRanking(item)
	}
	return db.Create(&rows).Error
}

func (r *rankingRepository) ListAll(ctx context.Context) ([]domain.Ranking, error) {
	var rows []rankingModel
	err := r.db.WithContext(ctx).Order("rank ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Ranking, len(rows))
	for i := range rows {
		out[i] = toRanking(&rows[i])
	}
	return out, nil
}

func (r *rankingRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]domain.Ranking, error) {
	var rows []rankingModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Ranking, len(rows))
	for i := range rows {
		out[i] = toRanking(&rows[i])
	}
	return out, nil
}

type refreshTokenRepository struct{ db *gorm.DB }

func (r *refreshTokenRepository) Create(ctx context.Context, tx domain.Transaction, token *domain.RefreshToken) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromRefreshToken(token)).Error
}

func (r *refreshTokenRepository) Update(ctx context.Context, tx domain.Transaction, token *domain.RefreshToken) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Save(fromRefreshToken(token)).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var m refreshTokenModel
	err := r.db.WithContext(ctx).First(&m, "token_hash = ?", tokenHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toRefreshToken(&m), nil
}

func (r *refreshTokenRepository) RevokeAllActiveByUserID(ctx context.Context, tx domain.Transaction, userID uuid.UUID, now time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Model(&refreshTokenModel{}).
		Where("user_id = ? AND status = ?", userID, domain.RefreshTokenActive).
		Updates(map[string]any{
			"status":     domain.RefreshTokenRevoked,
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

func (r *refreshTokenRepository) RevokeFamily(ctx context.Context, tx domain.Transaction, familyID uuid.UUID, now time.Time) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Model(&refreshTokenModel{}).
		Where("family_id = ? AND status = ?", familyID, domain.RefreshTokenActive).
		Updates(map[string]any{
			"status":     domain.RefreshTokenRevoked,
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("(status = ? AND expires_at < ?) OR (status = ? AND revoked_at < ?)",
			domain.RefreshTokenActive, before,
			domain.RefreshTokenRevoked, before).
		Delete(&refreshTokenModel{})
	return res.RowsAffected, res.Error
}

type activityLogRepository struct{ db *gorm.DB }

func (r *activityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	return r.db.WithContext(ctx).Create(fromActivityLog(log)).Error
}

func (r *activityLogRepository) Search(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]domain.ActivityLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&activityLogModel{})
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
	var rows []activityLogModel
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.ActivityLog, len(rows))
	for i := range rows {
		out[i] = toActivityLog(&rows[i])
	}
	return out, total, nil
}

func (r *activityLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&activityLogModel{})
	return res.RowsAffected, res.Error
}

func (r *activityLogRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]domain.ActivityLog, error) {
	var rows []activityLogModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.ActivityLog, len(rows))
	for i := range rows {
		out[i] = toActivityLog(&rows[i])
	}
	return out, nil
}

type loginLogRepository struct{ db *gorm.DB }

func (r *loginLogRepository) Create(ctx context.Context, tx domain.Transaction, log *domain.LoginLog) error {
	db, err := conn(ctx, r.db, tx)
	if err != nil {
		return err
	}
	return db.Create(fromLoginLog(log)).Error
}

func (r *loginLogRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.LoginLog, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&loginLogModel{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []loginLogModel
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.LoginLog, len(rows))
	for i := range rows {
		out[i] = toLoginLog(&rows[i])
	}
	return out, total, nil
}

func (r *loginLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Limit(batchSize).
		Delete(&loginLogModel{})
	return res.RowsAffected, res.Error
}

func (r *loginLogRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]domain.LoginLog, error) {
	var rows []loginLogModel
	err := r.db.WithContext(ctx).Where("updated_at > ?", since).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.LoginLog, len(rows))
	for i := range rows {
		out[i] = toLoginLog(&rows[i])
	}
	return out, nil
}

type wsConnectionLogRepository struct{ db *gorm.DB }

func (r *wsConnectionLogRepository) Create(ctx context.Context, log *domain.WSConnectionLog) error {
	return r.db.WithContext(ctx).Create(fromWSConnectionLog(log)).Error
}

func (r *wsConnectionLogRepository) UpdateDisconnected(ctx context.Context, id uuid.UUID, disconnectedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&wsConnectionLogModel{}).
		Where("id = ?", id).
		Update("disconnected_at", disconnectedAt).Error
}

func (r *wsConnectionLogRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WSConnectionLog, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&wsConnectionLogModel{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []wsConnectionLogModel
	offset := (page - 1) * limit
	if err := q.Order("connected_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.WSConnectionLog, len(rows))
	for i := range rows {
		out[i] = toWSConnectionLog(&rows[i])
	}
	return out, total, nil
}

func (r *wsConnectionLogRepository) DeleteOlderThan(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("connected_at < ?", before).
		Limit(batchSize).
		Delete(&wsConnectionLogModel{})
	return res.RowsAffected, res.Error
}
