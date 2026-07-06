package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type WSAuthController struct {
	WSAuth usecase.IWSAuthUsecase
}

func NewWSAuthController(wsAuth usecase.IWSAuthUsecase) *WSAuthController {
	return &WSAuthController{WSAuth: wsAuth}
}

func (h *WSAuthController) IssueTicket(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	ticket, err := h.WSAuth.IssueTicket(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, map[string]string{"ticket": ticket})
}
