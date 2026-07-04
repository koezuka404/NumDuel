package dto_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/usecase"
)

func TestWriteErrorConflictCodes(t *testing.T) {
	tests := []struct {
		err    error
		code   string
		status int
	}{
		{usecase.ErrUserInActiveGame, "user_in_active_game", http.StatusConflict},
		{usecase.ErrGameAlreadyFinished, "game_already_finished", http.StatusConflict},
		{usecase.ErrNotYourTurn, "not_your_turn", http.StatusConflict},
		{usecase.ErrGameAlreadyStarted, "game_already_started", http.StatusConflict},
		{usecase.ErrCannotDeleteSelf, "cannot_delete_self", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
			if err := dto.WriteError(c, tt.err); err != nil {
				t.Fatalf("WriteError: %v", err)
			}
			if rec.Code != tt.status {
				t.Fatalf("status %d want %d", rec.Code, tt.status)
			}
			var body struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("json: %v", err)
			}
			if body.Error.Code != tt.code {
				t.Fatalf("code %q want %q", body.Error.Code, tt.code)
			}
		})
	}
}
