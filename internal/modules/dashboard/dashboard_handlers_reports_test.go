//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_.*
package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"time-tracker/internal/modules/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleReports
func TestDashboardHandlers_HandleReports(t *testing.T) {
	SetAppDir()

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	task := &Task{
		ID:          1,
		UserID:      1,
		Title:       "Test Task",
		Description: "Test Task Description",
		Color:       "#FF5733",
		SortOrder:   1,
		IsCompleted: false,
	}
	reportRow := ReportRow{
		Task: task,
		DailyDurations: map[time.Time]time.Duration{
			now.Truncate(24 * time.Hour): 2 * time.Hour,
		},
		TotalDuration:   2 * time.Hour,
		DurationPercent: 100.0,
	}

	reportData := ReportData{
		ReportRows: []ReportRow{reportRow},
		Days:       []time.Time{now.Truncate(24 * time.Hour)},
		DailyTotalDuration: map[time.Time]time.Duration{
			now.Truncate(24 * time.Hour): 2 * time.Hour,
		},
		TotalDuration: 2 * time.Hour,
	}

	t.Run("RedirectToLoginWhenUserIsNotAuthenticated", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports", nil)

		handler.HandleReports(w, r)

		assert.Equal(t, http.StatusSeeOther, w.Result().StatusCode)
		assert.Contains(t, w.Result().Header.Get("Location"), "/login")
	})

	t.Run("RenderReportsPageWithValidData", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

		repo.On("Reports", user.ID, mock.Anything, mock.Anything, mock.Anything).Return(reportData)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports?month=2024-01", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleReports(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Color)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Title)
		assert.Contains(t, w.Body.String(), "Reports")
		assert.Contains(t, w.Body.String(), "2024-01")
	})

	t.Run("RenderReportsForPreviousMonth", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

		repo.On("Reports", user.ID, mock.Anything, mock.Anything, mock.Anything).Return(reportData)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports?month=2023-12", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleReports(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Color)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Title)
		assert.Contains(t, w.Body.String(), "Reports")
		assert.Contains(t, w.Body.String(), "2023-12")
	})

	t.Run("RenderHxRequestWithoutLayout", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

		repo.On("Reports", user.ID, mock.Anything, mock.Anything, mock.Anything).Return(reportData)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports?month=2024-01", nil)
		r.Header.Set("HX-Request", "true")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleReports(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Color)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Title)
		assert.NotContains(t, w.Body.String(), "<html>")
	})

	t.Run("InvalidMonthParameter", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

		repo.On("Reports", user.ID, mock.Anything, mock.Anything, mock.Anything).Return(reportData)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports?month=invalid", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleReports(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Color)
		assert.Contains(t, w.Body.String(), reportData.ReportRows[0].Task.Title)
		assert.Contains(t, w.Body.String(), "Reports")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandeReports
func TestDashboardHandlers_getMonthInterval(t *testing.T) {
	nowWithTimezone := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("ValidMonthStr", func(t *testing.T) {
		monthStr := "2023-12"
		expectedStart := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC)

		start, end := getMonthInterval(monthStr, nowWithTimezone)

		assert.Equal(t, expectedStart, start, "Start interval should match for valid monthStr")
		assert.Equal(t, expectedEnd, end, "End interval should match for valid monthStr")
	})

	t.Run("EmptyMonthStr", func(t *testing.T) {
		monthStr := ""
		expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)

		start, end := getMonthInterval(monthStr, nowWithTimezone)

		assert.Equal(t, expectedStart, start, "Start interval should match for empty monthStr")
		assert.Equal(t, expectedEnd, end, "End interval should match for empty monthStr")
	})

	t.Run("InvalidMonthStr", func(t *testing.T) {
		monthStr := "invalid-date"
		expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)

		start, end := getMonthInterval(monthStr, nowWithTimezone)

		assert.Equal(t, expectedStart, start, "Start interval should match for invalid monthStr")
		assert.Equal(t, expectedEnd, end, "End interval should match for invalid monthStr")
	})
}
