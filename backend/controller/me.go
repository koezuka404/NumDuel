package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type MeController struct {
	Auth    usecase.IAuthUsecase
	Profile usecase.IProfileUsecase
}

func NewMeController(auth usecase.IAuthUsecase, profile usecase.IProfileUsecase) *MeController {
	return &MeController{Auth: auth, Profile: profile}
}

func (h *MeController) Get(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Auth.GetMe(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, getMeResponse(out))
}

func (h *MeController) GetProfile(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Profile.GetProfile(c.Request().Context(), auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, getProfileResponse(out))
}

func (h *MeController) MatchHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := h.Profile.GetMatchHistory(c.Request().Context(), auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, matchHistoryResponse(items), page, limit, total)
}

func (h *MeController) LoginHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := h.Profile.GetLoginHistory(c.Request().Context(), auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, loginHistoryResponse(items), page, limit, total)
}

func (h *MeController) WSHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := h.Profile.GetWSHistory(c.Request().Context(), auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, wsHistoryResponse(items), page, limit, total)
}
