//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_.*
package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsNew
func TestDashboardHandlers_HandleRecordsNew(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/new", nil)

		handler.HandleRecordsNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("RenderRecordFormWithTaskAndTimes", func(t *testing.T) {
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

	t.Run("RenderRecordFormWithTask", func(t *testing.T) {
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

	t.Run("TaskNotFound", func(t *testing.T) {
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

	t.Run("AccessDenied", func(t *testing.T) {
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

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsCreate
func TestDashboardHandlers_HandleRecordsCreate(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", nil)

		handler.HandleRecordsCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("Success", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
		}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		repo.On("TaskByID", 1).Return(task)
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{})
		repo.On("CreateRecord", mock.Anything).Return(1, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "ok")
	})

	t.Run("ValidateIntersectingRecordsError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
			Color:       "#FF5733",
			SortOrder:   1,
			IsCompleted: false,
		}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		timeEnd1 := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		timeEnd2 := time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC)
		intersectingRecords := []*Record{
			{
				ID:        1,
				TaskID:    1,
				TimeStart: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
				TimeEnd:   &timeEnd1,
				Comment:   "Overlapping task",
				Task: &Task{
					ID:     1,
					Title:  "Test Task 1",
					UserID: 1,
				},
			},
			{
				ID:        2,
				TaskID:    2,
				TimeStart: time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
				TimeEnd:   &timeEnd2,
				Comment:   "Another overlapping task",
				Task: &Task{
					ID:     2,
					Title:  "Test Task 2",
					UserID: 1,
				},
			},
		}

		repo.On("Tasks", user.ID, "").Return([]*Task{task})
		repo.On("TaskByID", 1).Return(task)
		repo.On("RecordsWithTasks", mock.Anything).Return(intersectingRecords)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "The selected time overlaps with other entries")
	})

	t.Run("ParseFormError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := BadRequestPost("/records")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("HasErrors", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
		}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"123"}, // Invalid time start
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Time Start is invalid")
	})

	t.Run("TaskNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		repo.On("TaskByID", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		task := &Task{ID: 1, UserID: 2} // Task belongs to another user
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsCreate(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsEdit
func TestDashboardHandlers_HandleRecordsEdit(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)

		handler.HandleRecordsEdit(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("RecordNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		repo.On("RecordByIDWithTask", 1).Return(nil, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsEdit(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Record not found")
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{ID: 1, UserID: 2} // Task belongs to another user
		record := &Record{ID: 1, TaskID: 1, Task: task}
		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsEdit(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})

	t.Run("RenderEditForm", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
			IsCompleted: true,
		}
		timeEnd := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		record := &Record{
			ID:        1,
			TaskID:    1,
			TimeStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:   &timeEnd,
			Comment:   "Test comment",
			Task:      task,
		}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsEdit(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "Test comment")
		assert.Contains(t, w.Body.String(), "2024-01-01T12:00")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsUpdate
func TestDashboardHandlers_HandleRecordsUpdate(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records/1", nil)
		r.SetPathValue("id", "1")

		repo.On("RecordByIDWithTask", 1).Return(nil, nil)

		handler.HandleRecordsUpdate(w, r)

		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("FormParsingError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := BadRequestPost("/records/1")
		r.SetPathValue("id", "1")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Task Description",
			IsCompleted: true,
		}
		timeEnd := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		record := &Record{
			ID:        1,
			TaskID:    1,
			TimeStart: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:   &timeEnd,
			Comment:   "Test comment",
			Task:      task,
		}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)

		handler.HandleRecordsUpdate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{ID: 1, UserID: 1, Title: "Test Task", IsCompleted: true}
		record := &Record{ID: 1, TaskID: 1, TimeStart: time.Now(), TimeEnd: nil, Comment: "Test comment", Task: task}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"invalid-time"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records/1", strings.NewReader(form.Encode()))
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Time Start is invalid")
	})

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{ID: 1, UserID: 1, Title: "Test Task"}
		record := &Record{ID: 1, TaskID: 1, TimeStart: time.Now(), TimeEnd: nil, Comment: "Test comment", Task: task}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Updated comment"},
		}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})
		repo.On("UpdateRecord", mock.Anything).Return(nil)
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records/1", strings.NewReader(form.Encode()))
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "ok")
	})

	t.Run("IntersectingRecordsError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{ID: 1, UserID: 1, Title: "Test Task"}
		record := &Record{ID: 1, TaskID: 1, TimeStart: time.Now(), TimeEnd: nil, Comment: "Test comment", Task: task}

		intersectingRecords := []*Record{
			{
				ID:        2,
				TaskID:    1,
				TimeStart: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
				TimeEnd:   parseTimeFromInput("2024-01-01T15:00"),
				Comment:   "Overlapping task",
				Task:      task,
			},
		}

		form := url.Values{
			"task_id":    {"1"},
			"time_start": {"2024-01-01T12:00"},
			"time_end":   {"2024-01-01T14:00"},
			"comment":    {"Test comment"},
		}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("Tasks", user.ID, "").Return([]*Task{task})
		repo.On("RecordsWithTasks", mock.Anything).Return(intersectingRecords)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/records/1", strings.NewReader(form.Encode()))
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "The selected time overlaps with other entries")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsDelete
func TestDashboardHandlers_HandleRecordsDelete(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/records/1", nil)
		r.SetPathValue("id", "1")

		repo.On("RecordByIDWithTask", 1).Return(nil, nil)

		handler.HandleRecordsDelete(w, r)

		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("RecordNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/records/1", nil)
		r.SetPathValue("id", "1")

		user := &users.User{ID: 1, TimeZone: "UTC"}
		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("RecordByIDWithTask", 1).Return(nil, nil)

		handler.HandleRecordsDelete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Record not found")
	})

	t.Run("Success", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		task := &Task{ID: 1, UserID: 1, Title: "Test Task"}
		record := &Record{ID: 1, TaskID: 1, TimeStart: time.Now(), TimeEnd: nil, Comment: "Test comment", Task: task}

		repo.On("RecordByIDWithTask", 1).Return(record, nil)
		repo.On("DeleteRecord", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/records/1", nil)
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsDelete(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "ok", w.Body.String())
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleRecordsList
func TestDashboardHandlers_HandleRecordsList(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records", nil)

		handler.HandleRecordsList(w, r)

		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("SuccessfulRender", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

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
		repo.On("DailyRecords", mock.Anything, mock.Anything).Return(dailyRecords, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "This is a test record")
		assert.Contains(t, w.Body.String(), "Test Task")
	})

	t.Run("NoRecordsFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

		repo.On("DailyRecords", mock.Anything, mock.Anything).Return([]DailyRecords{}, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("WithWeekParameter", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}

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
		repo.On("DailyRecords", mock.Anything, mock.Anything).Return(dailyRecords, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records?week=2024-W01", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleRecordsList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "This is a test record")
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "2024-W01")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_getUserAndRecord
func TestDashboardHandlers_getUserAndRecord(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)

		handler.getUserAndRecord(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("InvalidRecordID", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		r := httptest.NewRequest(http.MethodGet, "/records/invalid-id", nil)
		r.SetPathValue("id", "invalid-id")
		ctx := context.WithValue(r.Context(), users.ContextUserKey, user)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.getUserAndRecord(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Invalid record ID")
	})

	t.Run("RecordNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")
		ctx := context.WithValue(r.Context(), users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("RecordByIDWithTask", 1).Return(nil)

		w := httptest.NewRecorder()
		handler.getUserAndRecord(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Record not found")
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")
		ctx := context.WithValue(r.Context(), users.ContextUserKey, user)
		r = r.WithContext(ctx)

		record := &Record{
			ID:     1,
			TaskID: 1,
			Task:   &Task{UserID: 2}, // Task belongs to another user
		}

		repo.On("RecordByIDWithTask", 1).Return(record)

		w := httptest.NewRecorder()
		handler.getUserAndRecord(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})

	t.Run("SuccessfulFetch", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		r := httptest.NewRequest(http.MethodGet, "/records/1", nil)
		r.SetPathValue("id", "1")
		ctx := context.WithValue(r.Context(), users.ContextUserKey, user)
		r = r.WithContext(ctx)

		record := &Record{
			ID:     1,
			TaskID: 1,
			Task:   &Task{UserID: 1}, // Task belongs to the current user
		}

		repo.On("RecordByIDWithTask", 1).Return(record)

		w := httptest.NewRecorder()
		fetchedUser, fetchedRecord := handler.getUserAndRecord(w, r)

		assert.Equal(t, user, fetchedUser)
		assert.Equal(t, record, fetchedRecord)
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_validateIntersectingRecords
func TestDashboardHandlers_validateIntersectingRecords(t *testing.T) {
	SetAppDir()

	t.Run("InvalidTimeEnd", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		form := recordForm{
			TimeStart: "2024-01-01T12:00",
			TimeEnd:   "2024-01-01T11:00", // TimeEnd less than TimeStart
		}
		formErrors := utils.FormErrors{}

		handler.validateIntersectingRecords(form, user, 0, formErrors)

		assert.Contains(t, formErrors, "TimeEnd")
		assert.Contains(t, formErrors["TimeEnd"], "Time End must be greater than Time Start")
	})

	t.Run("IntersectingTaskInProgress", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		form := recordForm{
			TimeStart: "2024-01-01T12:00",
			TimeEnd:   "2024-01-01T14:00", // Valid time range
		}
		formErrors := utils.FormErrors{}

		// Mock repository to return an ongoing task
		intersectingRecord := &Record{
			ID:        2,
			TaskID:    1,
			TimeStart: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			TimeEnd:   parseTimeFromInput("2024-01-01T15:00"),
			Comment:   "Ongoing task",
			Task:      &Task{ID: 1, UserID: 1, Title: "Ongoing Task"},
		}
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{intersectingRecord})

		handler.validateIntersectingRecords(form, user, 0, formErrors)

		assert.Contains(t, formErrors, "TimeEnd")
		assert.Contains(t, formErrors["TimeEnd"][0], "The selected time overlaps with other entries:")
	})

	t.Run("NoIntersectingRecordsInProgress", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		form := recordForm{
			TimeStart: "2024-01-01T12:00",
			TimeEnd:   "2024-01-01T14:00", // Valid time range
		}
		formErrors := utils.FormErrors{}

		// Mock repository to return no intersecting records
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{})

		handler.validateIntersectingRecords(form, user, 0, formErrors)

		assert.NotContains(t, formErrors, "TimeEnd") // No errors expected
	})

	t.Run("IntersectingRecords", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		form := recordForm{
			TimeStart: "2024-01-01T12:00",
			TimeEnd:   "2024-01-01T14:00", // Valid time range
		}
		formErrors := utils.FormErrors{}

		// Mock repository to return intersecting records
		intersectingRecord1 := &Record{
			ID:        2,
			TaskID:    1,
			TimeStart: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			TimeEnd:   parseTimeFromInput("2024-01-01T15:00"),
			Comment:   "Intersecting task 1",
			Task:      &Task{ID: 1, UserID: 1, Title: "Task 1"},
		}
		intersectingRecord2 := &Record{
			ID:        3,
			TaskID:    2,
			TimeStart: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			TimeEnd:   parseTimeFromInput("2024-01-01T16:00"),
			Comment:   "Intersecting task 2",
			Task:      &Task{ID: 2, UserID: 1, Title: "Task 2"},
		}
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{intersectingRecord1, intersectingRecord2})

		handler.validateIntersectingRecords(form, user, 0, formErrors)

		assert.Contains(t, formErrors, "TimeEnd")
		assert.Contains(t, formErrors["TimeEnd"][0], "The selected time overlaps with other entries:")
		assert.Contains(t, formErrors["TimeEnd"][0], "Intersecting task 1")
		assert.Contains(t, formErrors["TimeEnd"][0], "Intersecting task 2")
	})

	t.Run("IntersectingTaskInProgressWithNilTimeEnd", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC"}
		form := recordForm{
			TimeStart: "2024-01-01T12:00",
			TimeEnd:   "", // TimeEnd is nil
		}
		formErrors := utils.FormErrors{}

		// Mock repository to return an ongoing task
		intersectingRecord := &Record{
			ID:        2,
			TaskID:    1,
			TimeStart: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			TimeEnd:   nil,
			Comment:   "Ongoing task",
			Task:      &Task{ID: 1, UserID: 1, Title: "Ongoing Task"},
		}
		repo.On("RecordsWithTasks", mock.Anything).Return([]*Record{intersectingRecord})

		handler.validateIntersectingRecords(form, user, 0, formErrors)

		assert.Contains(t, formErrors, "TimeEnd")
		assert.Contains(t, formErrors["TimeEnd"][0], "You are already doing task:")
		assert.Contains(t, formErrors["TimeEnd"][0], "Ongoing task")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_parseTimeFromInput
func TestDashboardHandlers_parseTimeFromInput(t *testing.T) {
	t.Run("ValidTime", func(t *testing.T) {
		input := "2024-01-01T12:00"
		expected := parseTimeFromInput("2024-01-01T12:00")
		result := parseTimeFromInput(input)
		if result == nil || result.String() != expected.String() {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("EmptyInput", func(t *testing.T) {
		input := ""
		expected := (*time.Time)(nil)
		result := parseTimeFromInput(input)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("InvalidTime", func(t *testing.T) {
		input := "invalid"
		expected := (*time.Time)(nil)
		result := parseTimeFromInput(input)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
