// API 共通の JSON レスポンス形式を提供する。
// 成功: { "data": ... }  失敗: { "error": { "code", "message" } }
package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/internal/domain"
)

func WriteData(c echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

// WriteError は DomainError の code を HTTP ステータスに変換して返す。
func WriteError(c echo.Context, err error) error {
	if de, ok := domain.IsDomainError(err); ok {
		return c.JSON(statusForCode(de.Code), map[string]any{
			"error": map[string]string{"code": de.Code, "message": de.Error()},
		})
	}
	return c.JSON(http.StatusInternalServerError, map[string]any{
		"error": map[string]string{"code": domain.CodeInternalError, "message": "internal server error"},
	})
}

func statusForCode(code string) int {
	switch code {
	case domain.CodeValidation:
		return http.StatusBadRequest // 400
	case domain.CodeUnauthorized:
		return http.StatusUnauthorized // 401
	case domain.CodeTokenExpired:
		return http.StatusNotFound // 404（クライアントは refresh を試行）
	case domain.CodeDuplicateUser:
		return http.StatusConflict // 409
	default:
		return http.StatusInternalServerError // 500
	}
}
