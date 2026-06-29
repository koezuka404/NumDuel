// ゲーム API の HTTP ハンドラ。
package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type GameController struct {
	Deps usecase.GameDeps
}

func NewGameController(deps usecase.GameDeps) *GameController {
	return &GameController{Deps: deps}
}

// Get GET /api/games/:id
func (h *GameController) Get(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return dto.WriteError(c, model.ErrValidation("invalid game id"))
	}
	out, err := usecase.GetGameState(c.Request().Context(), h.Deps, auth.UserID, gameID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.GameStateToMap(out))
}
