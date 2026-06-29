package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/dto"
)

// Admin は role=master の JWT のみ通す（仕様 11.6）。
func Admin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, ok := AuthFrom(c)
			if !ok {
				return dto.WriteError(c, model.ErrUnauthorized())
			}
			if auth.Role != model.RoleMaster {
				return dto.WriteError(c, model.ErrForbidden("master role required"))
			}
			return next(c)
		}
	}
}
