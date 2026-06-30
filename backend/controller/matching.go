package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type MatchingController struct {
	Matching usecase.IMatchingUsecase
}

func NewMatchingController(matching usecase.IMatchingUsecase) *MatchingController {
	return &MatchingController{Matching: matching}
}

func (h *MatchingController) Start(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Matching.Start(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, startMatchingResponse(out))
}

func (h *MatchingController) Cancel(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Matching.Cancel(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, cancelMatchingResponse(out))
}

func (h *MatchingController) Status(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Matching.Status(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, matchingStatusResponse(out))
}
