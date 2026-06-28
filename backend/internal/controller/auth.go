// 認証 API の HTTP ハンドラ。JSON ↔ UseCase の変換と Cookie 操作を担当。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
	"github.com/numduel/numduel/internal/httputil"
	"github.com/numduel/numduel/internal/usecase"
)

const refreshCookieName = "refresh_token"

type AuthController struct {
	Auth                   usecase.AuthDeps
	CookieSecure           bool
	RefreshTokenExpiryDays int
}

func NewAuthController(auth usecase.AuthDeps, cookieSecure bool, refreshDays int) *AuthController {
	return &AuthController{Auth: auth, CookieSecure: cookieSecure, RefreshTokenExpiryDays: refreshDays}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register POST /api/auth/register — 新規ユーザー登録（JWT は発行しない）
func (h *AuthController) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, domain.ErrValidation("invalid request body"))
	}
	out, err := usecase.RegisterUser(c.Request().Context(), h.Auth, usecase.RegisterUserInput{
		Username: req.Username, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return httputil.WriteError(c, err)
	}
	return httputil.WriteData(c, http.StatusCreated, map[string]any{
		"id": out.ID.String(), "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	})
}

// Login POST /api/auth/login — JWT を Body、refresh を HttpOnly Cookie で返す
func (h *AuthController) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, domain.ErrValidation("invalid request body"))
	}
	out, err := usecase.Login(c.Request().Context(), h.Auth, usecase.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return httputil.WriteError(c, err)
	}
	setRefreshCookie(c, out.RefreshToken, h.RefreshTokenExpiryDays*86400, h.CookieSecure)
	return httputil.WriteData(c, http.StatusOK, map[string]string{"accessToken": out.AccessToken})
}

// Refresh POST /api/auth/refresh — Cookie の refresh_token で accessToken を更新
func (h *AuthController) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	out, err := usecase.RefreshToken(c.Request().Context(), h.Auth, usecase.RefreshTokenInput{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		return httputil.WriteError(c, err)
	}
	setRefreshCookie(c, out.RefreshToken, h.RefreshTokenExpiryDays*86400, h.CookieSecure)
	return httputil.WriteData(c, http.StatusOK, map[string]string{"accessToken": out.AccessToken})
}

// Logout POST /api/auth/logout — JWT 失効 + refresh 全失効 + Cookie 削除
func (h *AuthController) Logout(c echo.Context) error {
	auth, ok := httputil.AuthFrom(c)
	if !ok {
		return httputil.WriteError(c, domain.ErrUnauthorized())
	}
	if err := usecase.Logout(c.Request().Context(), h.Auth, usecase.LogoutInput{
		UserID: auth.UserID, JTI: auth.JTI, Exp: auth.ExpiresAt,
	}); err != nil {
		return httputil.WriteError(c, err)
	}
	clearRefreshCookie(c, h.CookieSecure)
	return c.NoContent(http.StatusNoContent)
}

func setRefreshCookie(c echo.Context, token string, maxAge int, secure bool) {
	c.SetCookie(&http.Cookie{
		Name: refreshCookieName, Value: token, Path: "/api/auth/refresh",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: maxAge,
	})
}

func clearRefreshCookie(c echo.Context, secure bool) {
	c.SetCookie(&http.Cookie{
		Name: refreshCookieName, Value: "", Path: "/api/auth/refresh",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: 0,
	})
}
