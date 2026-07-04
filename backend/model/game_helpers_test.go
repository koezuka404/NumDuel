package model_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestGameIsCurrentTurn(t *testing.T) {
	player := uuid.New()
	other := uuid.New()
	game := model.Game{CurrentTurnPlayerID: &player}
	if !game.IsCurrentTurn(player) || game.IsCurrentTurn(other) {
		t.Fatal("IsCurrentTurn with player set")
	}
	noTurn := model.Game{}
	if noTurn.IsCurrentTurn(player) {
		t.Fatal("nil turn player should be false")
	}
}

func TestGuessAndMatchHistoryTableNames(t *testing.T) {
	if (model.Guess{}).TableName() != "guesses" {
		t.Fatal("Guess TableName")
	}
	if (model.MatchHistory{}).TableName() != "match_histories" {
		t.Fatal("MatchHistory TableName")
	}
}
