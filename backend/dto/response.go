// API 共通の JSON レスポンス形式を提供する
// 成功: { "data": ... } 失敗: { "error": { "code", "message" } }
package dto

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
)

func WriteData(c echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

// WriteError は DomainError の code を HTTP ステータスに変換して返す
func WriteError(c echo.Context, err error) error {
	if de, ok := model.IsDomainError(err); ok {
		return c.JSON(statusForCode(de.Code), map[string]any{
			"error": map[string]string{"code": de.Code, "message": de.Error()},
		})
	}
	return c.JSON(http.StatusInternalServerError, map[string]any{
		"error": map[string]string{"code": model.CodeInternalError, "message": "internal server error"},
	})
}

func statusForCode(code string) int {
	switch code {
	case model.CodeValidation:
		return http.StatusBadRequest // 400
	case model.CodeUnauthorized:
		return http.StatusUnauthorized // 401
	case model.CodeTokenExpired:
		return http.StatusNotFound // 404（クライアントは refresh を試行）
	case model.CodeDuplicateUser, model.CodeUserInActiveGame, model.CodeAlreadyInMatching,
		model.CodeGameNotStarted, model.CodeGameAlreadyFinished, model.CodeNotYourTurn,
		model.CodeGameAlreadyStarted, model.CodeUserAlreadyDeleted:
		return http.StatusConflict // 409
	case model.CodeForbidden, model.CodeCannotDeleteSelf, model.CodeCannotDeleteMaster:
		return http.StatusForbidden // 403
	case model.CodeNotFound:
		return http.StatusNotFound // 404
	case model.CodeRateLimitExceeded:
		return http.StatusTooManyRequests // 429
	case model.CodeInvalidDigitLength, model.CodeInvalidDigit, model.CodeDuplicateDigit:
		return http.StatusBadRequest // 400
	default:
		return http.StatusInternalServerError // 500
	}
}
