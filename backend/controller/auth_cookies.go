package controller

import "net/http"

// newAuthCookie builds a cross-site auth cookie.
// SameSite=None is required because the frontend (numduel.onrender.com) and
// backend (numduel-backend.onrender.com) are on different origins with no
// same-origin proxy/rewrite in front of them. Secure=true is always set, so
// the cookie is only ever sent over HTTPS, and CORS is locked down to an
// explicit allow-list (see middleware.CORS) with AllowCredentials=true, which
// mitigates the CSRF exposure that SameSite=None would otherwise widen.
func newAuthCookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{ // #nosec G124 -- SameSite=None is intentional; see comment above
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   maxAge,
	}
}
