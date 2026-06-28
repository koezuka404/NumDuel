// ログインユーザー向け API。Middleware で検証済みの JWT 情報を使う。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/httputil"
	"github.com/numduel/numduel/internal/usecase"
)

type MeController struct {
	Auth usecase.AuthDeps
}

func NewMeController(auth usecase.AuthDeps) *MeController {
	return &MeController{Auth: auth}
}

// Get GET /api/me — 自分のプロフィール概要
func (h *MeController) Get(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	out, err := usecase.GetMe(c.Request().Context(), h.Auth, auth.UserID)
	if err != nil {
		return httputil.WriteError(c, err)
	}
	return httputil.WriteData(c, http.StatusOK, map[string]any{
		"id": out.ID.String(), "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	})
}
