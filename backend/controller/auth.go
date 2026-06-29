// 認証 API の HTTP ハンドラ。JSON ↔ UseCase の変換と Cookie 操作を担当。
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type AuthController struct {
	Auth                   usecase.AuthDeps
	CookieSecure           bool
	JWTExpiryMinutes       int
	RefreshTokenExpiryDays int
}

func NewAuthController(auth usecase.AuthDeps, cookieSecure bool, jwtMinutes, refreshDays int) *AuthController {
	return &AuthController{
		Auth: auth, CookieSecure: cookieSecure,
		JWTExpiryMinutes: jwtMinutes, RefreshTokenExpiryDays: refreshDays,
	}
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

// Register POST /api/auth/register
func (h *AuthController) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return dto.WriteError(c, model.ErrValidation("invalid request body"))
	}
	out, err := usecase.RegisterUser(c.Request().Context(), h.Auth, usecase.RegisterUserInput{
		Username: req.Username, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusCreated, usecase.RegisterUserResponse(out))
}

// Login POST /api/auth/login — access_token / refresh_token を HttpOnly Cookie で返す
func (h *AuthController) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return dto.WriteError(c, model.ErrValidation("invalid request body"))
	}
	out, err := usecase.Login(c.Request().Context(), h.Auth, usecase.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.setAuthCookies(c, out.AccessToken, out.RefreshToken)
	return dto.WriteData(c, http.StatusOK, usecase.LoginResponse(out))
}

// Refresh POST /api/auth/refresh
func (h *AuthController) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(middleware.RefreshCookieName)
	if err != nil || cookie.Value == "" {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	out, err := usecase.RefreshToken(c.Request().Context(), h.Auth, usecase.RefreshTokenInput{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.setAuthCookies(c, out.AccessToken, out.RefreshToken)
	return dto.WriteData(c, http.StatusOK, map[string]any{})
}

// Logout POST /api/auth/logout
func (h *AuthController) Logout(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	if err := usecase.Logout(c.Request().Context(), h.Auth, usecase.LogoutInput{
		UserID: auth.UserID, JTI: auth.JTI, Exp: auth.ExpiresAt,
	}); err != nil {
		return dto.WriteError(c, err)
	}
	h.clearAuthCookies(c)
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthController) setAuthCookies(c echo.Context, accessToken, refreshToken string) {
	c.SetCookie(&http.Cookie{
		Name: middleware.AccessCookieName, Value: accessToken, Path: "/",
		HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode,
		MaxAge: h.JWTExpiryMinutes * 60,
	})
	c.SetCookie(&http.Cookie{
		Name: middleware.RefreshCookieName, Value: refreshToken, Path: "/api/auth/refresh",
		HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode,
		MaxAge: h.RefreshTokenExpiryDays * 86400,
	})
}

func (h *AuthController) clearAuthCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name: middleware.AccessCookieName, Value: "", Path: "/",
		HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: 0,
	})
	c.SetCookie(&http.Cookie{
		Name: middleware.RefreshCookieName, Value: "", Path: "/api/auth/refresh",
		HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: 0,
	})
}
