package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

// §18.4.4 Game entity state transitions (via game_logic helpers + model.Game)
func TestGameSecretAndStart(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusWaitingSecret,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := setGameSecretHash(game, p1, "hash1"); err != nil {
		t.Fatalf("set player1 secret: %v", err)
	}
	if game.Status != model.GameStatusWaitingSecret || game.Player1Secret == "" {
		t.Fatal("expected WAITING_SECRET with player1 secret set")
	}
	if err := setGameSecretHash(game, p2, "hash2"); err != nil {
		t.Fatalf("set player2 secret: %v", err)
	}
	if err := startGame(game, now); err != nil {
		t.Fatalf("startGame: %v", err)
	}
	if game.Status != model.GameStatusInProgress {
		t.Fatalf("status = %s, want IN_PROGRESS", game.Status)
	}
	if game.CurrentTurn != 1 {
		t.Fatalf("currentTurn = %d, want 1", game.CurrentTurn)
	}
	if game.CurrentTurnPlayerID == nil || *game.CurrentTurnPlayerID != p1 {
		t.Fatal("expected player1 as first turn")
	}
}

func TestAddGuessTurnAdvance(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2,
		CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	results := JudgeDigits([4]int{5, 6, 7, 8}, [4]int{1, 2, 3, 4})
	guess, err := addGameGuess(game, p1, [4]int{1, 2, 3, 4}, results, false, now)
	if err != nil {
		t.Fatalf("addGameGuess: %v", err)
	}
	if guess.HitCount != 0 || game.CurrentTurn != 2 {
		t.Fatalf("turn=%d hitCount=%d", game.CurrentTurn, guess.HitCount)
	}
	if game.CurrentTurnPlayerID == nil || *game.CurrentTurnPlayerID != p2 {
		t.Fatal("expected turn passed to player2")
	}
}

func TestGamePlayerSlot(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{Player1ID: p1, Player2ID: p2, CreatedAt: now, UpdatedAt: now}
	slot, err := gamePlayerSlot(game, p1)
	if err != nil || slot != 1 {
		t.Fatalf("player1 slot: slot=%d err=%v", slot, err)
	}
	slot, err = gamePlayerSlot(game, p2)
	if err != nil || slot != 2 {
		t.Fatalf("player2 slot: slot=%d err=%v", slot, err)
	}
	if _, err := gamePlayerSlot(game, uuid.New()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("forbidden slot: %v", err)
	}
}

func TestSetGameSecretDuplicateAndForbidden(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{Player1ID: p1, Player2ID: p2, Player1Secret: "h1", CreatedAt: now, UpdatedAt: now}
	if err := setGameSecretHash(game, p1, "h2"); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("duplicate secret: %v", err)
	}
	if err := setGameSecretHash(game, uuid.New(), "h3"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("forbidden secret: %v", err)
	}
	_ = p2
}

func TestStartGameInvalidState(t *testing.T) {
	now := time.Now().UTC()
	game := &model.Game{Status: model.GameStatusWaitingSecret, CreatedAt: now, UpdatedAt: now}
	if err := startGame(game, now); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("start without secrets: %v", err)
	}
}

func TestGameOpponentHelpers(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		Player1ID: p1, Player2ID: p2,
		Player1Secret: "s1", Player2Secret: "s2",
		CreatedAt: now, UpdatedAt: now,
	}
	opp, err := gameOpponentID(game, p1)
	if err != nil || opp != p2 {
		t.Fatalf("opponent id: %v", err)
	}
	hash, slot, err := gameOpponentSecretHash(game, p1)
	if err != nil || hash != "s2" || slot != 2 {
		t.Fatalf("opponent secret p1: hash=%s slot=%d err=%v", hash, slot, err)
	}
	hash, slot, err = gameOpponentSecretHash(game, p2)
	if err != nil || hash != "s1" || slot != 1 {
		t.Fatalf("opponent secret p2: hash=%s slot=%d err=%v", hash, slot, err)
	}
	if _, err := gameOpponentID(game, uuid.New()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("opponent forbidden: %v", err)
	}
	if _, _, err := gameOpponentSecretHash(game, uuid.New()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("secret forbidden: %v", err)
	}
}

func TestAddGameGuessErrors(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	waiting := &model.Game{Status: model.GameStatusWaitingSecret, Player1ID: p1, Player2ID: p2, CreatedAt: now, UpdatedAt: now}
	results := JudgeDigits([4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8})
	if _, err := addGameGuess(waiting, p1, [4]int{1, 2, 3, 4}, results, false, now); !errors.Is(err, ErrGameNotStarted) {
		t.Fatalf("not started: %v", err)
	}
	inProgress := &model.Game{
		Status: model.GameStatusInProgress, Player1ID: p1, Player2ID: p2,
		CurrentTurn: 1, CurrentTurnPlayerID: &p1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := addGameGuess(inProgress, p2, [4]int{1, 2, 3, 4}, results, false, now); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("not your turn: %v", err)
	}
}

func TestAddGameGuessOpponentIDError(t *testing.T) {
	now := time.Now().UTC()
	outsider := uuid.New()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2,
		CurrentTurnPlayerID: &outsider, CurrentTurn: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	results := JudgeDigits([4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8})
	_, err := addGameGuess(game, outsider, [4]int{1, 2, 3, 4}, results, false, now)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("opponent id error: %v", err)
	}
}

func TestFinishGameErrors(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	finished := &model.Game{Status: model.GameStatusFinished, Player1ID: p1, Player2ID: p2, CreatedAt: now, UpdatedAt: now}
	if err := finishGame(finished, p1, now); !errors.Is(err, ErrGameAlreadyFinished) {
		t.Fatalf("already finished: %v", err)
	}
	inProgress := &model.Game{Status: model.GameStatusInProgress, Player1ID: p1, Player2ID: p2, CreatedAt: now, UpdatedAt: now}
	if err := finishGame(inProgress, uuid.New(), now); err == nil {
		t.Fatalf("invalid winner should fail")
	}
}

func TestCancelGameBySecretTimeoutWrongStatus(t *testing.T) {
	now := time.Now().UTC()
	game := &model.Game{Status: model.GameStatusInProgress, CreatedAt: now, UpdatedAt: now}
	if err := cancelGameBySecretTimeout(game, now); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("wrong status: %v", err)
	}
}

func TestFinishGame(t *testing.T) {
	now := time.Now().UTC()
	p1, p2 := uuid.New(), uuid.New()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusInProgress,
		Player1ID: p1, Player2ID: p2,
		CurrentTurn: 1, CurrentTurnPlayerID: &p1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := finishGame(game, p2, now); err != nil {
		t.Fatalf("finishGame: %v", err)
	}
	if game.Status != model.GameStatusFinished || game.WinnerID == nil || *game.WinnerID != p2 {
		t.Fatal("expected FINISHED with winner player2")
	}
}
