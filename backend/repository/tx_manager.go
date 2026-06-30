package repository

import (
	"context"

	"gorm.io/gorm"
)

// ITxRepos はトランザクション内で使う Repository 群を Usecase へ渡す窓口
type ITxRepos interface {
	Users() IUserRepository
	Games() IGameRepository
	Guesses() IGuessRepository
	MatchHistories() IMatchHistoryRepository
	MatchingQueue() IMatchingQueueRepository
	Rankings() IRankingRepository
	RefreshTokens() IRefreshTokenRepository
	LoginLogs() ILoginLogRepository
}

// TxManager は複数 DB 更新を 1 トランザクションとして実行する
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx ITxRepos) error) error
}

// GormTxManager は GORM の transaction 開始・commit・rollback を担当する
type GormTxManager struct {
	db *gorm.DB
}

type gormTxRepos struct {
	users          *userRepository
	games          *gameRepository
	guesses        *guessRepository
	matchHistories *matchHistoryRepository
	matchingQueue  *matchingQueueRepository
	rankings       *rankingRepository
	refreshTokens  IRefreshTokenRepository
	loginLogs      *loginLogRepository
}

// NewTxManager は TxManager を作成する
func NewTxManager(database *gorm.DB) TxManager {
	return &GormTxManager{db: database}
}

// WithinTx は fn を GORM transaction 内で実行するfn が error を返した場合は rollback する
func (m *GormTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx ITxRepos) error) error {
	return m.db.WithContext(ctx).Transaction(func(txDB *gorm.DB) error {
		txRepos := &gormTxRepos{
			users:          &userRepository{db: txDB},
			games:          &gameRepository{db: txDB},
			guesses:        &guessRepository{db: txDB},
			matchHistories: &matchHistoryRepository{db: txDB},
			matchingQueue:  &matchingQueueRepository{db: txDB},
			rankings:       &rankingRepository{db: txDB},
			refreshTokens:  NewRefreshTokenRepository(txDB),
			loginLogs:      &loginLogRepository{db: txDB},
		}
		return fn(ctx, txRepos)
	})
}

func (r *gormTxRepos) Users() IUserRepository                  { return r.users }
func (r *gormTxRepos) Games() IGameRepository                  { return r.games }
func (r *gormTxRepos) Guesses() IGuessRepository               { return r.guesses }
func (r *gormTxRepos) MatchHistories() IMatchHistoryRepository { return r.matchHistories }
func (r *gormTxRepos) MatchingQueue() IMatchingQueueRepository { return r.matchingQueue }
func (r *gormTxRepos) Rankings() IRankingRepository            { return r.rankings }
func (r *gormTxRepos) RefreshTokens() IRefreshTokenRepository  { return r.refreshTokens }
func (r *gormTxRepos) LoginLogs() ILoginLogRepository          { return r.loginLogs }
