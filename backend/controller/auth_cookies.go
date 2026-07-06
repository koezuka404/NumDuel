package controller

import (
	"net/http"
)

// newAuthCookie builds a same-site auth cookie (Vercel rewrites proxy → same origin).
func newAuthCookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}
