// ログインユーザー向け API。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type MeController struct {
	Auth    usecase.AuthDeps
	Profile usecase.ProfileDeps
}

func NewMeController(auth usecase.AuthDeps, profile usecase.ProfileDeps) *MeController {
	return &MeController{Auth: auth, Profile: profile}
}
func (h *MeController) Get(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.GetMe(c.Request().Context(), h.Auth, auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.GetMeResponse(out))
}

// GetProfile GET /api/me/profile
func (h *MeController) GetProfile(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.GetProfile(c.Request().Context(), h.Profile, auth.UserID)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.GetProfileResponse(out))
}

// MatchHistory GET /api/me/match-history
func (h *MeController) MatchHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := usecase.GetMatchHistory(c.Request().Context(), h.Profile, auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, usecase.MatchHistoryResponse(items), page, limit, total)
}

// LoginHistory GET /api/me/login-history
func (h *MeController) LoginHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := usecase.GetLoginHistory(c.Request().Context(), h.Profile, auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, usecase.LoginHistoryResponse(items), page, limit, total)
}

// WSHistory GET /api/me/ws-history
func (h *MeController) WSHistory(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := usecase.GetWSHistory(c.Request().Context(), h.Profile, auth.UserID, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, usecase.WSHistoryResponse(items), page, limit, total)
}
