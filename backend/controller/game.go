package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type GameController struct {
	Game usecase.IGameUsecase
}

func NewGameController(game usecase.IGameUsecase) *GameController {
	return &GameController{Game: game}
}

func (h *GameController) Get(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return dto.WriteError(c, usecase.ErrBadRequest)
	}
	out, err := h.Game.GetGameState(c.Request().Context(), auth.UserID, gameID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, gameStateResponse(out))
}
