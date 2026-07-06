package redis

import "github.com/numduel/numduel/usecase"

// nil *Store を interface に代入すると interface != nil になりメソッド呼び出しで panic する。
// 本番 Render で REDIS_URL / APP_ENV 未設定のとき auth 保護 API が 500 になる原因だった。

func JWTRevoker(s *Store) usecase.IJWTRevoker {
	if s == nil {
		return nil
	}
	return s
}

func ForceLogoutStore(s *Store) usecase.IForceLogoutStore {
	if s == nil {
		return nil
	}
	return s
}

func WSSessionStore(s *Store) usecase.IWSSessionStore {
	if s == nil {
		return nil
	}
	return s
}

func GameLockStore(s *Store) usecase.IGameLockStore {
	if s == nil {
		return nil
	}
	return s
}

func TurnStore(s *Store) usecase.ITurnStore {
	if s == nil {
		return nil
	}
	return s
}

func DistributedLockStore(s *Store) usecase.IDistributedLockStore {
	if s == nil {
		return nil
	}
	return s
}

func BackupStatusStore(s *Store) usecase.IBackupStatusStore {
	if s == nil {
		return nil
	}
	return s
}

func WSTicketStore(s *Store) usecase.IWSTicketStore {
	if s == nil {
		return nil
	}
	return s
}
