package repository

import (
	"github.com/numduel/numduel/model"
)

// model.Repository の GORM 実装。テーブルごとにサブリポジトリを持つ。
type Repository struct {
	users            *userRepository
	games            *gameRepository
	guesses          *guessRepository
	matchHistories   *matchHistoryRepository
	matchingQueue    *matchingQueueRepository
	rankings         *rankingRepository
	refreshTokens    IRefreshTokenRepository
	activityLogs     *activityLogRepository
	loginLogs        *loginLogRepository
	wsConnectionLogs *wsConnectionLogRepository
}

var _ model.Repository = (*Repository)(nil)

func NewRepository(database *DB) *Repository {
	g := database.Gorm()
	return &Repository{
		users:            &userRepository{db: g},
		games:            &gameRepository{db: g},
		guesses:          &guessRepository{db: g},
		matchHistories:   &matchHistoryRepository{db: g},
		matchingQueue:    &matchingQueueRepository{db: g},
		rankings:         &rankingRepository{db: g},
		refreshTokens:    NewRefreshTokenRepository(g),
		activityLogs:     &activityLogRepository{db: g},
		loginLogs:        &loginLogRepository{db: g},
		wsConnectionLogs: &wsConnectionLogRepository{db: g},
	}
}

func (r *Repository) Users() model.UserRepository                  { return r.users }
func (r *Repository) Games() model.GameRepository                  { return r.games }
func (r *Repository) Guesses() model.GuessRepository               { return r.guesses }
func (r *Repository) MatchHistories() model.MatchHistoryRepository { return r.matchHistories }
func (r *Repository) MatchingQueue() model.MatchingQueueRepository { return r.matchingQueue }
func (r *Repository) Rankings() model.RankingRepository            { return r.rankings }
func (r *Repository) RefreshTokens() model.RefreshTokenRepository  { return r.refreshTokens }
func (r *Repository) ActivityLogs() model.ActivityLogRepository    { return r.activityLogs }
func (r *Repository) LoginLogs() model.LoginLogRepository          { return r.loginLogs }
func (r *Repository) WSConnectionLogs() model.WSConnectionLogRepository {
	return r.wsConnectionLogs
}
