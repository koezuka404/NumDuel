// マッチング API の HTTP ハンドラ。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type MatchingController struct {
	Deps usecase.MatchingDeps
}

func NewMatchingController(deps usecase.MatchingDeps) *MatchingController {
	return &MatchingController{Deps: deps}
}

// Start POST /api/matching/start
func (h *MatchingController) Start(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.StartMatching(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.StartMatchingResponse(out))
}

// Cancel POST /api/matching/cancel
func (h *MatchingController) Cancel(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.CancelMatching(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.CancelMatchingResponse(out))
}

// Status GET /api/matching/status
func (h *MatchingController) Status(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.GetMatchingStatus(c.Request().Context(), h.Deps, auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.MatchingStatusResponse(out))
}
