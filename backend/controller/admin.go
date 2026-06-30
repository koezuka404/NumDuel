package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type AdminController struct {
	Admin usecase.IAdminUsecase
}

func NewAdminController(admin usecase.IAdminUsecase) *AdminController {
	return &AdminController{Admin: admin}
}

func (h *AdminController) ListUsers(c echo.Context) error {
	page, limit := dto.ParsePageLimit(c)
	items, total, err := h.Admin.ListUsers(c.Request().Context(), page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, adminUsersResponse(items), page, limit, total)
}

func (h *AdminController) SearchUsers(c echo.Context) error {
	items, err := h.Admin.SearchUsers(c.Request().Context(), c.QueryParam("q"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, adminUsersResponse(items))
}

func (h *AdminController) DeleteUser(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	targetID, err := parseUUIDParam(c, "id")
	if err != nil {
		return dto.WriteError(c, err)
	}
	if err := h.Admin.DeleteUser(c.Request().Context(), auth.UserID, targetID); err != nil {
		return dto.WriteError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminController) RebuildRanking(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	if err := h.Admin.RebuildRanking(c.Request().Context(), auth.UserID); err != nil {
		return dto.WriteError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminController) SearchLogs(c echo.Context) error {
	userID, err := parseOptionalUUID(c.QueryParam("userId"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	from, err := parseOptionalTime(c.QueryParam("from"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	to, err := parseOptionalTime(c.QueryParam("to"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	page, limit := dto.ParsePageLimit(c)
	items, total, err := h.Admin.SearchActivityLogs(c.Request().Context(), c.QueryParam("logType"), userID, from, to, page, limit)
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WritePaged(c, http.StatusOK, activityLogsResponse(items), page, limit, total)
}

func (h *AdminController) ListLogTypes(c echo.Context) error {
	types, err := h.Admin.ListActivityLogTypes(c.Request().Context())
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, logTypesResponse(types))
}

func (h *AdminController) DownloadLogs(c echo.Context) error {
	auth, ok := middleware.AuthFrom(c)
	if !ok {
		return dto.WriteError(c, usecase.ErrUnauthorized)
	}
	userID, err := parseOptionalUUID(c.QueryParam("userId"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	from, err := parseOptionalTime(c.QueryParam("from"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	to, err := parseOptionalTime(c.QueryParam("to"))
	if err != nil {
		return dto.WriteError(c, err)
	}
	csvData, err := h.Admin.DownloadActivityLogsCSV(c.Request().Context(), auth.UserID, c.QueryParam("logType"), userID, from, to)
	if err != nil {
		return dto.WriteError(c, err)
	}
	filename := fmt.Sprintf("activity_logs_%s.csv", time.Now().UTC().Format("20060102T150405Z"))
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv; charset=utf-8")
	return c.Blob(http.StatusOK, "text/csv", csvData)
}

func (h *AdminController) BackupStatus(c echo.Context) error {
	out, err := h.Admin.GetBackupStatus(c.Request().Context())
	if err != nil {
		return dto.WriteError(c, err)
	}
	return dto.WriteData(c, http.StatusOK, backupStatusResponse(out))
}

func parseUUIDParam(c echo.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, usecase.ErrBadRequest
	}
	return id, nil
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, usecase.ErrBadRequest
	}
	return &id, nil
}

func parseOptionalTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, usecase.ErrBadRequest
	}
	return &t, nil
}
