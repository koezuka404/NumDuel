package dto_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func TestWriteErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"unauthorized", usecase.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{"token expired", usecase.ErrTokenExpired, http.StatusNotFound, "token_expired"},
		{"duplicate user", usecase.ErrDuplicateUser, http.StatusConflict, "duplicate_user"},
		{"cannot delete master", usecase.ErrCannotDeleteMaster, http.StatusForbidden, "cannot_delete_master"},
		{"forbidden", usecase.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"not found", usecase.ErrNotFound, http.StatusNotFound, "not_found"},
		{"already in matching", usecase.ErrAlreadyInMatching, http.StatusConflict, "already_in_matching"},
		{"game not started", usecase.ErrGameNotStarted, http.StatusConflict, "game_not_started"},
		{"user already deleted", usecase.ErrUserAlreadyDeleted, http.StatusConflict, "user_already_deleted"},
		{"rate limit", usecase.ErrRateLimitExceeded, http.StatusTooManyRequests, "rate_limit_exceeded"},
		{"bad request", usecase.ErrBadRequest, http.StatusBadRequest, "validation_error"},
		{"validation", model.ErrBadUsername, http.StatusBadRequest, "validation_error"},
		{"internal", errors.New("boom"), http.StatusInternalServerError, "internal_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if err := dto.WriteError(c, tt.err); err != nil {
				t.Fatalf("WriteError: %v", err)
			}
			if rec.Code != tt.status {
				t.Fatalf("status %d want %d body=%s", rec.Code, tt.status, rec.Body.String())
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

func TestWriteData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := dto.WriteData(c, http.StatusOK, map[string]string{"status": "ok"}); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Data["status"] != "ok" {
		t.Fatalf("data: %+v", body.Data)
	}
}
