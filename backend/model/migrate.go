package model

func MigrateTargets() []any {
	return []any{
		&User{},
		&Game{},
		&Guess{},
		&MatchHistory{},
		&Ranking{},
		&MatchingQueueEntry{},
		&ActivityLog{},
		&LoginLog{},
		&WSConnectionLog{},
		&RefreshToken{},
	}
}
