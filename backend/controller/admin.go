package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type AdminController struct {
	Deps usecase.AdminDeps
}

func NewAdminController(deps usecase.AdminDeps) *AdminController {
	return &AdminController{Deps: deps}
}

// ListUsers GET /api/admin/users
func (h *AdminController) ListUsers(c echo.Context) error {
	page, limit := dto.ParsePageLimit(c)
	items, total, err := usecase.GetAdminUsers(c.Request().Context(), h.Deps, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, usecase.AdminUsersResponse(items), page, limit, total)
}

// SearchUsers GET /api/admin/users/search
func (h *AdminController) SearchUsers(c echo.Context) error {
	items, err := usecase.SearchAdminUsers(c.Request().Context(), h.Deps, c.QueryParam("q"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.AdminUsersResponse(items))
}

// DeleteUser DELETE /api/admin/users/:id
func (h *AdminController) DeleteUser(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	targetID, err := parseUUIDParam(c, "id")
	if err != nil {
		return dto.WriteError(c, err)
	}
	if err := usecase.DeleteUser(c.Request().Context(), h.Deps, auth.UserID, targetID); err != nil {
		return dto.WriteError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// RebuildRanking POST /api/admin/ranking/rebuild
func (h *AdminController) RebuildRanking(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, model.ErrUnauthorized())
	}
	if err := usecase.RebuildRankingAsAdmin(c.Request().Context(), h.Deps, auth.UserID); err != nil {
		return dto.WriteError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// SearchLogs GET /api/admin/logs
func (h *AdminController) SearchLogs(c echo.Context) error {
	userID, err := usecase.ParseOptionalUUID(c.QueryParam("userId"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	from, err := usecase.ParseOptionalTime(c.QueryParam("from"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	to, err := usecase.ParseOptionalTime(c.QueryParam("to"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := usecase.SearchActivityLogs(c.Request().Context(), h.Deps, c.QueryParam("logType"), userID, from, to, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, usecase.ActivityLogsResponse(items), page, limit, total)
}

// ListLogTypes GET /api/admin/logs/types
func (h *AdminController) ListLogTypes(c echo.Context) error {
	types, err := usecase.ListActivityLogTypes(c.Request().Context(), h.Deps)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.LogTypesResponse(types))
}

// DownloadLogs GET /api/admin/logs/download
func (h *AdminController) DownloadLogs(c echo.Context) error {
	userID, err := usecase.ParseOptionalUUID(c.QueryParam("userId"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	from, err := usecase.ParseOptionalTime(c.QueryParam("from"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	to, err := usecase.ParseOptionalTime(c.QueryParam("to"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	csvData, err := usecase.DownloadActivityLogsCSV(c.Request().Context(), h.Deps, c.QueryParam("logType"), userID, from, to)
	if err != nil {
		return dto.WriteError(c, err)
	}
	filename := fmt.Sprintf("activity_logs_%s.csv", time.Now().UTC().Format("20060102T150405Z"))
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv; charset=utf-8")
	return c.Blob(http.StatusOK, "text/csv", csvData)
}

// BackupStatus GET /api/admin/backup/status
func (h *AdminController) BackupStatus(c echo.Context) error {
	out, err := usecase.GetBackupStatus(c.Request().Context(), h.Deps)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, usecase.BackupStatusResponse(out))
}

func parseUUIDParam(c echo.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, model.ErrValidation("invalid id")
	}
	return id, nil
}
