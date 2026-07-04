package controller_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/controller"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type stubAuthUC struct {
	logoutErr error
	getMeErr  error
}

func (s *stubAuthUC) Register(context.Context, usecase.RegisterInput) (*usecase.RegisterResult, error) {
	return nil, nil
}
func (s *stubAuthUC) Login(context.Context, usecase.LoginInput) (*usecase.LoginResult, error) {
	return nil, nil
}
func (s *stubAuthUC) Refresh(context.Context, usecase.RefreshInput) (*usecase.RefreshResult, error) {
	return nil, nil
}
func (s *stubAuthUC) Logout(_ context.Context, _ usecase.LogoutInput) error {
	return s.logoutErr
}
func (s *stubAuthUC) GetMe(_ context.Context, userID uuid.UUID) (*usecase.MeResult, error) {
	if s.getMeErr != nil {
		return nil, s.getMeErr
	}
	return &usecase.MeResult{ID: userID.String(), Username: "alice", Role: string(model.RoleUser)}, nil
}
func (s *stubAuthUC) SeedMaster(context.Context, usecase.SeedMasterInput) error { return nil }
func (s *stubAuthUC) CleanupExpiredRefreshTokens(context.Context)               {}

type stubProfileUC struct {
	profileErr      error
	matchHistoryErr error
	loginHistoryErr error
	wsHistoryErr    error
}

func (s *stubProfileUC) GetProfile(_ context.Context, _ uuid.UUID) (*usecase.GetProfileOutput, error) {
	if s.profileErr != nil {
		return nil, s.profileErr
	}
	return &usecase.GetProfileOutput{Username: "alice", WinCount: 0}, nil
}
func (s *stubProfileUC) GetMatchHistory(context.Context, uuid.UUID, int, int) ([]usecase.MatchHistoryItem, int64, error) {
	if s.matchHistoryErr != nil {
		return nil, 0, s.matchHistoryErr
	}
	return nil, 0, nil
}
func (s *stubProfileUC) GetLoginHistory(context.Context, uuid.UUID, int, int) ([]usecase.LoginHistoryItem, int64, error) {
	if s.loginHistoryErr != nil {
		return nil, 0, s.loginHistoryErr
	}
	return nil, 0, nil
}
func (s *stubProfileUC) GetWSHistory(context.Context, uuid.UUID, int, int) ([]usecase.WSConnectionHistoryItem, int64, error) {
	if s.wsHistoryErr != nil {
		return nil, 0, s.wsHistoryErr
	}
	return nil, 0, nil
}

type stubMatchingUC struct {
	startErr  error
	cancelErr error
	statusErr error
}

func (s *stubMatchingUC) Start(context.Context, uuid.UUID) (*usecase.StartMatchingOutput, error) {
	if s.startErr != nil {
		return nil, s.startErr
	}
	return &usecase.StartMatchingOutput{Status: "waiting"}, nil
}
func (s *stubMatchingUC) Cancel(context.Context, uuid.UUID) (*usecase.CancelMatchingOutput, error) {
	if s.cancelErr != nil {
		return nil, s.cancelErr
	}
	return &usecase.CancelMatchingOutput{Status: "cancelled"}, nil
}
func (s *stubMatchingUC) Status(context.Context, uuid.UUID) (*usecase.GetMatchingStatusOutput, error) {
	if s.statusErr != nil {
		return nil, s.statusErr
	}
	return &usecase.GetMatchingStatusOutput{Status: "idle"}, nil
}

type stubAdminUC struct {
	listUsersErr   error
	searchUsersErr error
	deleteUserErr  error
	rebuildErr     error
	searchLogsErr  error
	listTypesErr   error
	downloadErr    error
	backupErr      error
}

func (s *stubAdminUC) ListUsers(context.Context, int, int) ([]usecase.AdminUserItem, int64, error) {
	if s.listUsersErr != nil {
		return nil, 0, s.listUsersErr
	}
	return nil, 0, nil
}
func (s *stubAdminUC) SearchUsers(context.Context, string) ([]usecase.AdminUserItem, error) {
	if s.searchUsersErr != nil {
		return nil, s.searchUsersErr
	}
	return []usecase.AdminUserItem{{ID: uuid.New(), Username: "alice"}}, nil
}
func (s *stubAdminUC) DeleteUser(context.Context, uuid.UUID, uuid.UUID) error {
	return s.deleteUserErr
}
func (s *stubAdminUC) SearchActivityLogs(context.Context, string, *uuid.UUID, *time.Time, *time.Time, int, int) ([]usecase.ActivityLogItem, int64, error) {
	if s.searchLogsErr != nil {
		return nil, 0, s.searchLogsErr
	}
	return nil, 0, nil
}
func (s *stubAdminUC) ListActivityLogTypes(context.Context) ([]string, error) {
	if s.listTypesErr != nil {
		return nil, s.listTypesErr
	}
	return []string{"login"}, nil
}
func (s *stubAdminUC) DownloadActivityLogsCSV(context.Context, uuid.UUID, string, *uuid.UUID, *time.Time, *time.Time) ([]byte, error) {
	if s.downloadErr != nil {
		return nil, s.downloadErr
	}
	return []byte("id\n"), nil
}
func (s *stubAdminUC) RebuildRanking(context.Context, uuid.UUID) error {
	return s.rebuildErr
}
func (s *stubAdminUC) GetBackupStatus(context.Context) (*usecase.BackupStatusOutput, error) {
	if s.backupErr != nil {
		return nil, s.backupErr
	}
	return &usecase.BackupStatusOutput{Status: "idle"}, nil
}

type stubRankingUC struct {
	getErr error
}

func (s *stubRankingUC) Get(context.Context) ([]usecase.RankingItem, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return nil, nil
}
func (s *stubRankingUC) Rebuild(context.Context) error   { return nil }
func (s *stubRankingUC) RunScheduledRebuild(context.Context) error { return nil }

type stubGameUC struct {
	getErr error
}

func (s *stubGameUC) GetGameState(context.Context, uuid.UUID, uuid.UUID) (*usecase.GameStateOutput, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return nil, nil
}
func (s *stubGameUC) SyncGameState(context.Context, uuid.UUID, uuid.UUID) (*usecase.GameStateOutput, error) {
	return nil, nil
}
func (s *stubGameUC) SetSecretNumber(context.Context, uuid.UUID, uuid.UUID, string) error { return nil }
func (s *stubGameUC) SubmitGuess(context.Context, uuid.UUID, uuid.UUID, string, bool) error {
	return nil
}
func (s *stubGameUC) HandleTimeout(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s *stubGameUC) CancelBySecretTimeout(context.Context, uuid.UUID) error     { return nil }
func (s *stubGameUC) RecoverActiveGames(context.Context) error                   { return nil }

func testAuthInfo(role model.Role) middleware.AuthInfo {
	return middleware.AuthInfo{
		UserID: uuid.New(), Role: role, JTI: "jti",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
}

func callHandler(t *testing.T, h echo.HandlerFunc, method, path string, auth *middleware.AuthInfo, params map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if len(params) > 0 {
		names := make([]string, 0, len(params))
		vals := make([]string, 0, len(params))
		for k, v := range params {
			names = append(names, k)
			vals = append(vals, v)
		}
		c.SetParamNames(names...)
		c.SetParamValues(vals...)
	}
	if auth != nil {
		middleware.SetAuth(c, *auth)
	}
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	return rec
}

func TestAuthLogoutUseCaseError(t *testing.T) {
	auth := controller.NewAuthController(&stubAuthUC{logoutErr: errors.New("logout failed")}, false, 60, 7)
	info := testAuthInfo(model.RoleUser)
	rec := callHandler(t, auth.Logout, http.MethodPost, "/logout", &info, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("logout usecase error status %d", rec.Code)
	}
}

func TestAuthLogoutDirectUnauthorized(t *testing.T) {
	auth := controller.NewAuthController(&stubAuthUC{}, false, 60, 7)
	rec := callHandler(t, auth.Logout, http.MethodPost, "/logout", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("logout direct unauthorized status %d", rec.Code)
	}
}

func TestMeUseCaseErrors(t *testing.T) {
	fail := errors.New("profile failed")
	me := controller.NewMeController(&stubAuthUC{getMeErr: fail}, &stubProfileUC{
		profileErr: fail, matchHistoryErr: fail, loginHistoryErr: fail, wsHistoryErr: fail,
	})
	info := testAuthInfo(model.RoleUser)

	cases := []struct {
		name string
		h    echo.HandlerFunc
	}{
		{"Get", me.Get},
		{"GetProfile", me.GetProfile},
		{"MatchHistory", me.MatchHistory},
		{"LoginHistory", me.LoginHistory},
		{"WSHistory", me.WSHistory},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := callHandler(t, tc.h, http.MethodGet, "/me", &info, nil)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("%s status %d", tc.name, rec.Code)
			}
		})
	}
}

func TestMatchingUseCaseErrors(t *testing.T) {
	fail := errors.New("matching failed")
	match := controller.NewMatchingController(&stubMatchingUC{
		startErr: fail, cancelErr: fail, statusErr: fail,
	})
	info := testAuthInfo(model.RoleUser)

	for name, h := range map[string]echo.HandlerFunc{
		"Start":  match.Start,
		"Cancel": match.Cancel,
		"Status": match.Status,
	} {
		t.Run(name, func(t *testing.T) {
			method := http.MethodPost
			if name == "Status" {
				method = http.MethodGet
			}
			rec := callHandler(t, h, method, "/matching", &info, nil)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("%s status %d", name, rec.Code)
			}
		})
	}
}

func TestMatchingHandlersWithoutAuth(t *testing.T) {
	match := controller.NewMatchingController(&stubMatchingUC{})
	for name, h := range map[string]echo.HandlerFunc{
		"Start":  match.Start,
		"Cancel": match.Cancel,
		"Status": match.Status,
	} {
		t.Run(name, func(t *testing.T) {
			method := http.MethodPost
			if name == "Status" {
				method = http.MethodGet
			}
			rec := callHandler(t, h, method, "/matching", nil, nil)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status %d", name, rec.Code)
			}
		})
	}
}

func TestAdminUseCaseErrors(t *testing.T) {
	fail := errors.New("admin failed")
	admin := controller.NewAdminController(&stubAdminUC{
		listUsersErr: fail, searchUsersErr: fail, deleteUserErr: fail,
		rebuildErr: fail, searchLogsErr: fail, listTypesErr: fail,
		downloadErr: fail, backupErr: fail,
	})
	info := testAuthInfo(model.RoleMaster)

	cases := []struct {
		name   string
		method string
		h      echo.HandlerFunc
		path   string
	}{
		{"ListUsers", http.MethodGet, admin.ListUsers, "/admin/users"},
		{"SearchUsers", http.MethodGet, admin.SearchUsers, "/admin/users/search?q=ali"},
		{"DeleteUser", http.MethodDelete, admin.DeleteUser, "/admin/users/x"},
		{"RebuildRanking", http.MethodPost, admin.RebuildRanking, "/admin/ranking/rebuild"},
		{"SearchLogs", http.MethodGet, admin.SearchLogs, "/admin/logs"},
		{"ListLogTypes", http.MethodGet, admin.ListLogTypes, "/admin/logs/types"},
		{"DownloadLogs", http.MethodGet, admin.DownloadLogs, "/admin/logs/download"},
		{"BackupStatus", http.MethodGet, admin.BackupStatus, "/admin/backup/status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]string{}
			if tc.name == "DeleteUser" {
				params["id"] = uuid.New().String()
			}
			rec := callHandler(t, tc.h, tc.method, tc.path, &info, params)
			if tc.name == "DownloadLogs" {
				if rec.Code != http.StatusInternalServerError {
					t.Fatalf("download status %d", rec.Code)
				}
				return
			}
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("%s status %d", tc.name, rec.Code)
			}
		})
	}
}

func TestAdminSearchUsersSuccess(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	info := testAuthInfo(model.RoleMaster)
	rec := callHandler(t, admin.SearchUsers, http.MethodGet, "/admin/users/search?q=ali", &info, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search users status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminDeleteUserDirectUnauthorized(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	rec := callHandler(t, admin.DeleteUser, http.MethodDelete, "/admin/users/x", nil, map[string]string{
		"id": uuid.New().String(),
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("delete unauthorized status %d", rec.Code)
	}
}

func TestAdminDownloadLogsDirectUnauthorized(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	rec := callHandler(t, admin.DownloadLogs, http.MethodGet, "/admin/logs/download", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("download unauthorized status %d", rec.Code)
	}
}

func TestAdminDownloadLogsSuccess(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	info := testAuthInfo(model.RoleMaster)
	rec := callHandler(t, admin.DownloadLogs, http.MethodGet, "/admin/logs/download", &info, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("download status %d", rec.Code)
	}
}

func TestAdminRebuildRankingDirectUnauthorized(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	rec := callHandler(t, admin.RebuildRanking, http.MethodPost, "/admin/ranking/rebuild", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("rebuild direct unauthorized status %d", rec.Code)
	}
}

func TestAdminDownloadLogsBadQueryDirect(t *testing.T) {
	admin := controller.NewAdminController(&stubAdminUC{})
	info := testAuthInfo(model.RoleMaster)
	for _, query := range []string{
		"/admin/logs/download?userId=bad-uuid",
		"/admin/logs/download?from=bad",
	} {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, query, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		middleware.SetAuth(c, info)
		if err := admin.DownloadLogs(c); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status %d", query, rec.Code)
		}
	}
}

func TestRankingGetUseCaseError(t *testing.T) {
	ranking := controller.NewRankingController(&stubRankingUC{getErr: errors.New("ranking failed")})
	rec := callHandler(t, ranking.Get, http.MethodGet, "/ranking", nil, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("ranking status %d", rec.Code)
	}
}

func TestGameGetUseCaseError(t *testing.T) {
	game := controller.NewGameController(&stubGameUC{getErr: errors.New("game failed")})
	info := testAuthInfo(model.RoleUser)
	rec := callHandler(t, game.Get, http.MethodGet, "/games/x", &info, map[string]string{
		"id": uuid.New().String(),
	})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("game status %d", rec.Code)
	}
}

func TestGameGetUnauthorized(t *testing.T) {
	game := controller.NewGameController(&stubGameUC{})
	rec := callHandler(t, game.Get, http.MethodGet, "/games/x", nil, map[string]string{
		"id": uuid.New().String(),
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("game unauthorized status %d", rec.Code)
	}
}

func TestMeHandlersWithoutAuth(t *testing.T) {
	me := controller.NewMeController(&stubAuthUC{}, &stubProfileUC{})
	for name, h := range map[string]echo.HandlerFunc{
		"Get":           me.Get,
		"GetProfile":    me.GetProfile,
		"MatchHistory":  me.MatchHistory,
		"LoginHistory":  me.LoginHistory,
		"WSHistory":     me.WSHistory,
	} {
		t.Run(name, func(t *testing.T) {
			rec := callHandler(t, h, http.MethodGet, "/me", nil, nil)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s status %d", name, rec.Code)
			}
		})
	}
}
