package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/infrastructure/postgres"
)

func setupTestRepo(t *testing.T) domain.Repository {
	t.Helper()
	db, err := postgres.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.CreateIndexes(); err != nil {
		t.Fatalf("indexes: %v", err)
	}
	return postgres.NewRepository(db)
}

func TestUserAndGameRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	user := &domain.User{
		ID:             uuid.New(),
		Username:       "alice",
		Email:          "alice@example.com",
		PasswordHash:   "hash",
		Role:           domain.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.Users().Create(ctx, nil, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := repo.Users().FindByEmail(ctx, user.Email)
	if err != nil || found == nil || found.Username != "alice" {
		t.Fatalf("find user: err=%v found=%v", err, found)
	}

	player2 := &domain.User{
		ID:             uuid.New(),
		Username:       "bob",
		Email:          "bob@example.com",
		PasswordHash:   "hash",
		Role:           domain.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.Users().Create(ctx, nil, player2); err != nil {
		t.Fatalf("create player2: %v", err)
	}

	game := &domain.Game{
		ID:          uuid.New(),
		Status:      domain.GameStatusWaitingSecret,
		Player1ID:   user.ID,
		Player2ID:   player2.ID,
		CurrentTurn: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.Games().Create(ctx, nil, game); err != nil {
		t.Fatalf("create game: %v", err)
	}

	active, err := repo.Games().FindActiveByUserID(ctx, user.ID)
	if err != nil || active == nil || active.ID != game.ID {
		t.Fatalf("find active game: err=%v active=%v", err, active)
	}
}

func TestGuessRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	gameID := uuid.New()
	playerID := uuid.New()
	guess := &domain.Guess{
		ID:           uuid.New(),
		GameID:       gameID,
		PlayerID:     playerID,
		Turn:         1,
		GuessNumber:  "1234",
		DigitResults: []domain.DigitResult{domain.DigitHit, domain.DigitMiss, domain.DigitMiss, domain.DigitMiss},
		HitCount:     1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Guesses().Create(ctx, nil, guess); err != nil {
		t.Fatalf("create guess: %v", err)
	}

	rows, err := repo.Guesses().ListByGameAndPlayer(ctx, gameID, playerID)
	if err != nil || len(rows) != 1 || rows[0].HitCount != 1 {
		t.Fatalf("list guesses: err=%v rows=%v", err, rows)
	}
}

func TestMatchingQueueRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	entry := &domain.MatchingQueueEntry{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Status:    domain.MatchingQueueWaiting,
		CreatedAt: now,
	}
	if err := repo.MatchingQueue().Insert(ctx, nil, entry); err != nil {
		t.Fatalf("insert queue: %v", err)
	}

	exists, err := repo.MatchingQueue().ExistsWaitingByUserID(ctx, entry.UserID)
	if err != nil || !exists {
		t.Fatalf("exists waiting: err=%v exists=%v", err, exists)
	}
}

func TestRankingRepositoryReplaceAll(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	items := []domain.Ranking{
		domain.NewRanking(uuid.New(), 1, "alice", 3, now),
		domain.NewRanking(uuid.New(), 2, "bob", 1, now),
	}
	if err := repo.Rankings().ReplaceAll(ctx, nil, items); err != nil {
		t.Fatalf("replace rankings: %v", err)
	}

	all, err := repo.Rankings().ListAll(ctx)
	if err != nil || len(all) != 2 || all[0].Rank != 1 {
		t.Fatalf("list rankings: err=%v all=%v", err, all)
	}
}

func TestRefreshTokenRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	token := domain.NewRefreshToken(uuid.New(), "deadbeef", uuid.New(), now.Add(24*time.Hour), now)
	if err := repo.RefreshTokens().Create(ctx, nil, &token); err != nil {
		t.Fatalf("create token: %v", err)
	}

	found, err := repo.RefreshTokens().FindByTokenHash(ctx, "deadbeef")
	if err != nil || found == nil || found.Status != domain.RefreshTokenActive {
		t.Fatalf("find token: err=%v found=%v", err, found)
	}
}

func TestActivityLogRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	userID := uuid.New()

	log := &domain.ActivityLog{
		ID:        uuid.New(),
		UserID:    &userID,
		LogType:   "guess",
		Detail:    json.RawMessage(`{"gameId":"test"}`),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.ActivityLogs().Create(ctx, log); err != nil {
		t.Fatalf("create activity log: %v", err)
	}

	items, total, err := repo.ActivityLogs().Search(ctx, "guess", &userID, nil, nil, 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("search logs: err=%v total=%d items=%v", err, total, items)
	}
}

func TestTransactionCommit(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	tx, err := repo.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	user := &domain.User{
		ID:             uuid.New(),
		Username:       "txuser",
		Email:          "tx@example.com",
		PasswordHash:   "hash",
		Role:           domain.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.Users().Create(ctx, tx, user); err != nil {
		_ = repo.Rollback(tx)
		t.Fatalf("create in tx: %v", err)
	}
	if err := repo.Commit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	found, err := repo.Users().FindByID(ctx, user.ID)
	if err != nil || found == nil {
		t.Fatalf("find after commit: err=%v found=%v", err, found)
	}
}
