package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

func openRepos(t *testing.T) repository.Repos {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	return repos
}

func newUser(username, email string) *model.User {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.User{
		ID:             uuid.New(),
		Username:       username,
		Email:          email,
		PasswordHash:   "hash",
		Role:           model.RoleUser,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func createUser(t *testing.T, repos repository.Repos, username, email string) *model.User {
	t.Helper()
	user := newUser(username, email)
	if err := repos.User.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func newGame(p1, p2 uuid.UUID, status model.GameStatus) *model.Game {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.Game{
		ID:          uuid.New(),
		Status:      status,
		Player1ID:   p1,
		Player2ID:   p2,
		CurrentTurn: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func newGuess(gameID, playerID uuid.UUID, turn int) *model.Guess {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.Guess{
		ID:           uuid.New(),
		GameID:       gameID,
		PlayerID:     playerID,
		Turn:         turn,
		GuessNumber:  "1234",
		DigitResults: []model.DigitResult{model.DigitHit, model.DigitMiss, model.DigitHit, model.DigitMiss},
		HitCount:     2,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func newActivityLog(logType string, userID *uuid.UUID) *model.ActivityLog {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.ActivityLog{
		ID:        uuid.New(),
		UserID:    userID,
		LogType:   logType,
		Detail:    json.RawMessage(`{"k":"v"}`),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newLoginLog(userID uuid.UUID, action model.LoginAction) *model.LoginLog {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.LoginLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newWSLog(userID uuid.UUID) *model.WSConnectionLog {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.WSConnectionLog{
		ID:           uuid.New(),
		UserID:       userID,
		ConnectionID: "conn-" + uuid.New().String()[:8],
		ConnectedAt:  now,
	}
}

func newMatchHistory(gameID, winnerID, loserID uuid.UUID) *model.MatchHistory {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &model.MatchHistory{
		ID:             uuid.New(),
		GameID:         gameID,
		WinnerID:       winnerID,
		LoserID:        loserID,
		WinnerUsername: "winner",
		LoserUsername:  "loser",
		FinishedAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
