// マッチング API の HTTP ハンドラ。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/httputil"
	"github.com/numduel/numduel/internal/usecase"
)

type MatchingController struct {
	Deps usecase.MatchingDeps
}

func NewMatchingController(deps usecase.MatchingDeps) *MatchingController {
	return &MatchingController{Deps: deps}
}

// Start POST /api/matching/start
func (h *MatchingController) Start(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	out, err := usecase.StartMatching(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return httputil.WriteError(c, err)
	}
	return httputil.WriteData(c, http.StatusOK, map[string]string{"status": out.Status})
}

// Cancel POST /api/matching/cancel
func (h *MatchingController) Cancel(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	out, err := usecase.CancelMatching(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return httputil.WriteError(c, err)
	}
	return httputil.WriteData(c, http.StatusOK, map[string]string{"status": out.Status})
}

// Status GET /api/matching/status
func (h *MatchingController) Status(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	out, err := usecase.GetMatchingStatus(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return httputil.WriteError(c, err)
	}
	data := map[string]any{"status": out.Status, "gameId": nil}
	if out.GameID != nil {
		data["gameId"] = out.GameID.String()
	}
	return httputil.WriteData(c, http.StatusOK, data)
}
