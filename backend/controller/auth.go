package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type AuthController struct {
	Auth                   usecase.IAuthUsecase
	CookieSecure           bool
	JWTExpiryMinutes       int
	RefreshTokenExpiryDays int
}

func NewAuthController(auth usecase.IAuthUsecase, cookieSecure bool, jwtMinutes, refreshDays int) *AuthController {
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

func (h *AuthController) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return dto.WriteError(c, usecase.ErrBadRequest)
	}
	out, err := h.Auth.Register(c.Request().Context(), usecase.RegisterInput{
		Username: req.Username, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusCreated, registerUserResponse(out))
}

func (h *AuthController) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return dto.WriteError(c, usecase.ErrBadRequest)
	}
	out, err := h.Auth.Login(c.Request().Context(), usecase.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.setAuthCookies(c, out.AccessToken, out.RefreshToken)
	return dto.WriteData(c, http.StatusOK, loginResponse(out))
}

func (h *AuthController) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(middleware.RefreshCookieName)
	if err != nil || cookie.Value == "" {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	out, err := h.Auth.Refresh(c.Request().Context(), usecase.RefreshInput{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		return dto.WriteError(c, err)
	}
	h.setAuthCookies(c, out.AccessToken, out.RefreshToken)
	return dto.WriteData(c, http.StatusOK, map[string]any{})
}

func (h *AuthController) Logout(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	if err := h.Auth.Logout(c.Request().Context(), usecase.LogoutInput{
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
