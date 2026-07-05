package controller

import (
	"net/http"
)

// newAuthCookie builds a cross-site auth cookie (Vercel frontend ↔ Render backend).
func newAuthCookie(name, value, path string, maxAge int) *http.Cookie {
	// #nosec G124 -- SameSite=None with Secure is required for cross-origin cookie auth
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   maxAge,
	}
}
