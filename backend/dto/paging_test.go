package dto_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
)

func TestParsePageLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=2&limit=200", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	page, limit := dto.ParsePageLimit(c)
	if page != 2 || limit != 100 {
		t.Fatalf("page=%d limit=%d", page, limit)
	}

	req = httptest.NewRequest(http.MethodGet, "/?page=0&limit=-1", nil)
	c = e.NewContext(req, httptest.NewRecorder())
	page, limit = dto.ParsePageLimit(c)
	if page != 1 || limit != 20 {
		t.Fatalf("defaults page=%d limit=%d", page, limit)
	}
}

func TestWritePaged(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := dto.WritePaged(c, http.StatusOK, []string{"a"}, 1, 20, 1); err != nil {
		t.Fatalf("WritePaged: %v", err)
	}
	var body struct {
		Data struct {
			Items []string `json:"items"`
			Total int64    `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(body.Data.Items) != 1 || body.Data.Total != 1 {
		t.Fatalf("body: %+v", body.Data)
	}
}
