package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/usecase"
)

type RankingController struct {
	Ranking usecase.IRankingUsecase
}

func NewRankingController(ranking usecase.IRankingUsecase) *RankingController {
	return &RankingController{Ranking: ranking}
}

func (h *RankingController) Get(c echo.Context) error {
	items, err := h.Ranking.Get(c.Request().Context())
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, rankingResponse(items))
}
