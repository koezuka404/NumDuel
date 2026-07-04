package usecase

import (
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestGameStateWSPayload(t *testing.T) {
	state := &GameStateOutput{
		GameID:              uuid.New(),
		Status:              model.GameStatusInProgress,
		CurrentTurn:         2,
		CurrentTurnPlayerID: uuid.New().String(),
		RemainingSeconds:    25,
		MyGuesses: []GuessSummary{
			{Turn: 1, GuessNumber: "9012", HitCount: 1, IsAuto: false},
		},
		OpponentGuessCount: 1,
	}
	payload := gameStateWSPayload(state)
	if payload["gameId"] == nil || payload["myGuesses"] == nil {
		t.Fatalf("payload: %+v", payload)
	}
	guesses, ok := payload["myGuesses"].([]map[string]any)
	if !ok || len(guesses) != 1 {
		t.Fatalf("myGuesses: %+v", payload["myGuesses"])
	}
}
