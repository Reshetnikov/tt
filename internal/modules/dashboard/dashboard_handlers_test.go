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

func TestDashboardHandlers_HandleDashboard(t *testing.T) {
	SetAppDir()

	t.Run("redirects to login if user is nil", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

		handler.HandleDashboard(w, r)

		assert.Equal(t, http.StatusSeeOther, w.Result().StatusCode)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("renders dashboard with tasks and records", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}

		tasks := []*Task{
			{
				ID:          1,
				UserID:      1,
				Title:       "Test Task",
				Description: "This is a test task",
				Color:       "#FF5733",
				SortOrder:   1,
				IsCompleted: false,
			},
		}
		dailyRecords := []DailyRecords{
			{
				Day: time.Now(),
				Records: []Record{
					{
						ID:                1,
						TaskID:            1,
						TimeStart:         time.Now().Add(-2 * time.Hour),
						TimeEnd:           nil,
						Comment:           "This is a test record",
						Task:              tasks[0],
						StartPercent:      0.0,
						DurationPercent:   10.0,
						Duration:          2 * time.Hour,
						TimeStartIntraday: time.Now().Add(-2 * time.Hour),
						TimeEndIntraday:   time.Now(),
					},
				},
			},
		}
		repo.On("Tasks", user.ID, "").Return(tasks)
		repo.On("DailyRecords", mock.Anything, mock.Anything).Return(dailyRecords)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleDashboard(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Tasks & Records Dashboard")
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "This is a test record")
	})
}

func TestDashboardHandlers_getWeekInterval(t *testing.T) {
	t.Run("returns week interval from weekStr", func(t *testing.T) {
		weekStr := "2024-W01"
		isWeekStartMonday := true
		start, end := getWeekInterval(weekStr, time.Now(), isWeekStartMonday)

		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), start)
		assert.Equal(t, time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), end)
	})

	t.Run("falls back to GetWeekIntervalByDate when weekStr is empty", func(t *testing.T) {
		isWeekStartMonday := false
		now := time.Now()
		start, end := getWeekInterval("", now, isWeekStartMonday)

		assert.NotZero(t, start)
		assert.NotZero(t, end)
		assert.True(t, start.Before(end))
	})

	t.Run("handles invalid weekStr by falling back to current week", func(t *testing.T) {
		weekStr := "invalid"
		isWeekStartMonday := true
		now := time.Now()
		start, end := getWeekInterval(weekStr, now, isWeekStartMonday)

		assert.NotZero(t, start)
		assert.NotZero(t, end)
		assert.True(t, start.Before(end))
	})
}
