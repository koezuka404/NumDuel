package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/usecase"
)

type RankingController struct {
	Deps usecase.RankingDeps
}

func NewRankingController(deps usecase.RankingDeps) *RankingController {
	return &RankingController{Deps: deps}
}

// Get GET /api/ranking — 上位 3 名
func (h *RankingController) Get(c echo.Context) error {
	items, err := usecase.GetRanking(c.Request().Context(), h.Deps)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.RankingResponse(items))
}
