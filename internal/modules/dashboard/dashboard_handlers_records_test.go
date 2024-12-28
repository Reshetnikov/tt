//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_.*
package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time-tracker/internal/modules/users"

	"github.com/stretchr/testify/assert"
)

func TestDashboardHandlers_HandleRecordsNew(t *testing.T) {
	SetAppDir()

	t.Run("redirects to login if user is nil", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new", nil)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("renders form with pre-filled task and times", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
			Color:       "#FF5733",
			SortOrder:   1,
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new?taskId=1&date=2024-01-01", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "2024-01-01T")
	})

	t.Run("renders form with pre-filled task", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
			Color:       "#FF5733",
			SortOrder:   1,
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new?taskId=1", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "value=\"\"")
	})

	t.Run("returns 404 if task does not exist", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}
		repo.On("TaskByID", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new?taskId=1", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task not found")
	})

	t.Run("returns 403 if user does not own the task", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}
		task := &Task{ID: 1, UserID: 2} // Task belongs to another user
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new?taskId=1", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})
}
