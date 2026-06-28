// ゲーム API の HTTP ハンドラ。
package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/httputil"
	"github.com/numduel/numduel/internal/usecase"
)

type GameController struct {
	Deps usecase.GameDeps
}

func NewGameController(deps usecase.GameDeps) *GameController {
	return &GameController{Deps: deps}
}

// Get GET /api/games/:id
func (h *GameController) Get(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httputil.WriteError(c, domain.ErrValidation("invalid game id"))
	}
	out, err := usecase.GetGameState(c.Request().Context(), h.Deps, auth.UserID, gameID)
	if err != nil {
		return httputil.WriteError(c, err)
	}
	return httputil.WriteData(c, http.StatusOK, usecase.GameStateToMap(out))
}
