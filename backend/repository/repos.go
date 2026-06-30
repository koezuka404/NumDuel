package repository

import "gorm.io/gorm"

type Repos struct {
	DB              *gorm.DB
	User            IUserRepo
	Game            IGameRepo
	Guess           IGuessRepo
	MatchHistory    IMatchHistoryRepo
	MatchingQueue   IMatchingQueueRepo
	Ranking         IRankingRepo
	RefreshToken    IRefreshTokenRepo
	ActivityLog     IActivityLogRepo
	LoginLog        ILoginLogRepo
	WSConnectionLog IWSConnectionLogRepo
}

func NewRepos(db *gorm.DB) Repos {
	return Repos{
		DB:              db,
		User:            NewUserRepo(db),
		Game:            NewGameRepo(db),
		Guess:           NewGuessRepo(db),
		MatchHistory:    NewMatchHistoryRepo(db),
		MatchingQueue:   NewMatchingQueueRepo(db),
		Ranking:         NewRankingRepo(db),
		RefreshToken:    NewRefreshTokenRepo(db),
		ActivityLog:     NewActivityLogRepo(db),
		LoginLog:        NewLoginLogRepo(db),
		WSConnectionLog: NewWSConnectionLogRepo(db),
	}
}
