package model_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestGameParticipantAndTurn(t *testing.T) {
	p1 := uuid.New()
	p2 := uuid.New()
	other := uuid.New()
	game := model.Game{
		Status:              model.GameStatusInProgress,
		Player1ID:           p1,
		Player2ID:           p2,
		Player1Secret:       "h1",
		Player2Secret:       "h2",
		CurrentTurnPlayerID: &p1,
	}

	if !game.IsParticipant(p1) || !game.IsParticipant(p2) || game.IsParticipant(other) {
		t.Fatalf("IsParticipant mismatch")
	}
	if !game.IsCurrentTurn(p1) || game.IsCurrentTurn(p2) {
		t.Fatalf("IsCurrentTurn mismatch")
	}
	if !game.CanGuess(p1) || game.CanGuess(p2) {
		t.Fatalf("CanGuess mismatch")
	}
	if !game.BothSecretsSet() {
		t.Fatalf("BothSecretsSet expected true")
	}
	if game.IsFinished() {
		t.Fatalf("IsFinished expected false")
	}

	game.Status = model.GameStatusFinished
	if !game.IsFinished() || game.CanGuess(p1) {
		t.Fatalf("finished game state mismatch")
	}
}

func TestGameWaitingSecret(t *testing.T) {
	p1 := uuid.New()
	game := model.Game{
		Status:    model.GameStatusWaitingSecret,
		Player1ID: p1,
		Player2ID: uuid.New(),
	}
	if game.BothSecretsSet() || game.CanGuess(p1) {
		t.Fatalf("waiting secret should not allow guess")
	}
}
