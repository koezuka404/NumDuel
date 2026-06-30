// panic を 500 internal_error JSON に変換しログ出力
package middleware

import (
	"errors"
	"log"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
)

// Recover は Handler / 下流 Middleware 内の panic を捕捉する
// クライアントには { error: { code: internal_error } } を返し、詳細はサーバログのみ
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic: %v", r)
					err = dto.WriteError(c, errors.New("internal server error"))
				}
			}()
			return next(c)
		}
	}
}
