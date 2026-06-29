package repository

// Repository は IRepository の GORM 実装。テーブルごとにサブリポジトリを持つ。
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

var _ IRepository = (*Repository)(nil)

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

func (r *Repository) Users() IUserRepository                  { return r.users }
func (r *Repository) Games() IGameRepository                  { return r.games }
func (r *Repository) Guesses() IGuessRepository               { return r.guesses }
func (r *Repository) MatchHistories() IMatchHistoryRepository { return r.matchHistories }
func (r *Repository) MatchingQueue() IMatchingQueueRepository { return r.matchingQueue }
func (r *Repository) Rankings() IRankingRepository            { return r.rankings }
func (r *Repository) RefreshTokens() IRefreshTokenRepository  { return r.refreshTokens }
func (r *Repository) ActivityLogs() IActivityLogRepository    { return r.activityLogs }
func (r *Repository) LoginLogs() ILoginLogRepository          { return r.loginLogs }
func (r *Repository) WSConnectionLogs() IWSConnectionLogRepository {
	return r.wsConnectionLogs
}
