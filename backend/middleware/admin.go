package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func Admin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, ok := AuthFrom(c)
			if !ok {
				return dto.WriteError(c, usecase.ErrUnauthorized)
			}
			if auth.Role != model.RoleMaster {
				return dto.WriteError(c, usecase.ErrForbidden)
			}
			return next(c)
		}
	}
}
