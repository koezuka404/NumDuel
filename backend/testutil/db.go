package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/db"
	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/usecase"
)

const (
	TestJWTSecret = "test-jwt-secret-key-at-least-32-chars!!"
	TestPepper    = "test-game-secret-pepper-32bytes-min!!"
)

// OpenSQLiteDB opens an in-memory SQLite DB and runs migrations (spec §18.3).
func OpenSQLiteDB(t *testing.T) (*gorm.DB, repository.Repos) {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb, repository.NewRepos(gdb)
}

func NewAuthUC(t *testing.T, repos repository.Repos) *usecase.AuthUseCase {
	t.Helper()
	return NewAuthUCWithRevoker(t, repos, nil)
}

func NewAuthUCWithRevoker(t *testing.T, repos repository.Repos, revoker usecase.IJWTRevoker) *usecase.AuthUseCase {
	t.Helper()
	jwtSvc, err := infrcrypto.NewJWTService(TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	return usecase.NewAuthUseCase(
		repos,
		infrcrypto.NewPasswordService(),
		jwtSvc,
		infrcrypto.NewRefreshTokenService(),
		revoker, nil,
		7, 7,
	)
}

func NewGameUC(t *testing.T, repos repository.Repos) *usecase.GameUseCase {
	t.Helper()
	hasher, err := infrcrypto.NewSecretHashService(TestPepper)
	if err != nil {
		t.Fatalf("secret hasher: %v", err)
	}
	return usecase.NewGameUseCase(
		repos, hasher, nil, nil,
		infrcrypto.NewRandomNumberService(), nil,
		30*time.Second, 60*time.Second, 2*time.Second,
	)
}

func NewMatchingUC(repos repository.Repos) *usecase.MatchingUseCase {
	return usecase.NewMatchingUseCase(repos, nil)
}

func NewRankingUC(repos repository.Repos) *usecase.RankingUseCase {
	return usecase.NewRankingUseCase(repos, nil, 5*time.Second)
}

func NewAdminUC(repos repository.Repos, ranking *usecase.RankingUseCase) *usecase.AdminUseCase {
	return usecase.NewAdminUseCase(repos, ranking, nil, nil, nil, nil, 5*time.Second)
}

func NewGameUCWithNotifier(t *testing.T, repos repository.Repos, notifier usecase.IEventNotifier) *usecase.GameUseCase {
	t.Helper()
	hasher, err := infrcrypto.NewSecretHashService(TestPepper)
	if err != nil {
		t.Fatalf("secret hasher: %v", err)
	}
	return usecase.NewGameUseCase(
		repos, hasher, nil, nil,
		infrcrypto.NewRandomNumberService(), notifier,
		30*time.Second, 60*time.Second, 2*time.Second,
	)
}

func SeedMaster(t *testing.T, repos repository.Repos, email, password string) *model.User {
	t.Helper()
	auth := NewAuthUC(t, repos)
	if err := auth.SeedMaster(t.Context(), usecase.SeedMasterInput{
		Email: email, Password: password,
	}); err != nil {
		t.Fatalf("seed master: %v", err)
	}
	user, err := repos.User.FindByUsername(t.Context(), "admin")
	if err != nil || user == nil {
		t.Fatalf("find master: %v", err)
	}
	return user
}

func CreateUser(t *testing.T, repos repository.Repos, username, email, password string) *model.User {
	t.Helper()
	auth := NewAuthUC(t, repos)
	out, err := auth.Register(t.Context(), usecase.RegisterInput{
		Username: username, Email: email, Password: password,
	})
	if err != nil {
		t.Fatalf("register %s: %v", username, err)
	}
	id, err := uuid.Parse(out.ID)
	if err != nil {
		t.Fatalf("parse id: %v", err)
	}
	user, err := repos.User.FindByID(t.Context(), id)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	return user
}
