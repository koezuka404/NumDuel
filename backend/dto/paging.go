package dto

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const defaultPage = 1
const defaultLimit = 20
const maxLimit = 100

// ParsePageLimit は page / limit クエリを正規化する（仕様 6.1.4）。
func ParsePageLimit(c echo.Context) (page, limit int) {
	page = defaultPage
	limit = defaultLimit
	if v, err := strconv.Atoi(c.QueryParam("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.QueryParam("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func WritePaged(c echo.Context, status int, items any, page, limit int, total int64) error {
	return WriteData(c, status, map[string]any{
		"items": items, "page": page, "limit": limit, "total": total,
	})
}
