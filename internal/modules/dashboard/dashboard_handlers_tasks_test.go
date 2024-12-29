//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_.*
package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time-tracker/internal/modules/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTasksNew
func TestDashboardHandlers_HandleTasksNew(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/new", nil)

		handler.HandleTasksNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("RenderTaskForm", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1, TimeZone: "UTC", IsWeekStartMonday: true}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/new", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksNew(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "value=\"#EEEEEE\"")
		assert.Contains(t, w.Body.String(), "/tasks")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTasksCreate
func TestDashboardHandlers_HandleTasksCreate(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/create", nil)

		handler.HandleTasksCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("ParseFormError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		w := httptest.NewRecorder()
		r := BadRequestPost("/tasks/create")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksCreate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Bad Request")
	})

	t.Run("RenderTaskFormWithErrors", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		form := url.Values{
			"title":       {"title test"},
			"description": {"description test"},
			"color":       {"colorInvalid"}, // Invalid
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/create", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Color is invalid") // Assuming this is added in templates for errors
		assert.Contains(t, w.Body.String(), "/tasks")
	})

	t.Run("CreateTaskSuccess", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		form := url.Values{
			"title":       {"title test"},
			"description": {"description test"},
			"color":       {"#DDAA88"}, // Invalid
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/create", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("CreateTask", mock.Anything).Return(1, nil)

		handler.HandleTasksCreate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "load-tasks, close-modal", w.Header().Get("HX-Trigger"))
		assert.Equal(t, "ok", w.Body.String())
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTasksEdit
func TestDashboardHandlers_HandleTasksEdit(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/edit/1", nil)
		r.SetPathValue("id", "1")

		handler.HandleTasksEdit(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("TaskNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		repo.On("TaskByID", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksEdit(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task not found")
	})

	t.Run("RenderTaskFormWithTask", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
			Color:       "#FF5733",
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksEdit(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Test Task")
		assert.Contains(t, w.Body.String(), "Test Description")
		assert.Contains(t, w.Body.String(), "#FF5733")
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:     1,
			UserID: 2, // Task belongs to another user
		}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksEdit(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTasksUpdate
func TestDashboardHandlers_HandleTasksUpdate(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("TaskNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		repo.On("TaskByID", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task not found")
	})

	t.Run("ParseFormError", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
			Color:       "#FF5733",
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)
		w := httptest.NewRecorder()
		r := BadRequestPost("/tasks/1")
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Bad Request")
	})

	t.Run("RenderTaskFormWithErrors", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
			Color:       "#FF5733",
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)

		form := url.Values{
			"title":       {"title test"},
			"description": {"description test"},
			"color":       {"invalidColor"}, // Invalid color
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", strings.NewReader(form.Encode()))
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Color is invalid")
	})

	t.Run("UpdateTaskSuccess", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
			Color:       "#FF5733",
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)
		repo.On("GetMaxSortOrder", mock.Anything, mock.Anything).Return(1)

		form := url.Values{
			"title":        {"Updated Task Title"},
			"description":  {"Updated Description"},
			"color":        {"#DDAA88"},
			"is_completed": {"true"},
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", strings.NewReader(form.Encode()))
		r.SetPathValue("id", "1")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("UpdateTask", mock.Anything).Return(nil)

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "load-tasks, load-records, close-modal", w.Header().Get("HX-Trigger"))
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:     1,
			UserID: 2, // Task belongs to another user
		}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksUpdate(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTasksDelete
func TestDashboardHandlers_HandleTasksDelete(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)

		handler.HandleTasksDelete(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("TaskNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		repo.On("TaskByID", 1).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksDelete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task not found")
	})

	t.Run("DeleteTaskSuccess", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
			Color:       "#FF5733",
			IsCompleted: false,
		}
		repo.On("TaskByID", 1).Return(task)
		repo.On("DeleteTask", task.ID).Return(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksDelete(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "load-tasks, load-records", w.Header().Get("HX-Trigger"))
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("AccessDenied", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{
			ID:     1,
			UserID: 2, // Task belongs to another user
		}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTasksDelete(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleTaskList
func TestDashboardHandlers_HandleTaskList(t *testing.T) {
	SetAppDir()

	t.Run("RenderBlockNeedLogin", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks", nil)

		handler.HandleTaskList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature. Please")
	})

	t.Run("RenderTaskListWithNoTasks", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		repo.On("Tasks", user.ID, "").Return([]*Task{})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTaskList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Create Task")
		assert.NotContains(t, w.Body.String(), "draggable")
	})

	t.Run("RenderTaskListWithTasks", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task1 := &Task{ID: 1, Title: "Task 1", UserID: 1, IsCompleted: false}
		task2 := &Task{ID: 2, Title: "Task 2", UserID: 1, IsCompleted: true}
		repo.On("Tasks", user.ID, "").Return([]*Task{task1, task2})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTaskList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task 1")
		assert.Contains(t, w.Body.String(), "Task 2")
	})

	t.Run("RenderTaskListWithCompletedTasks", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task1 := &Task{ID: 1, Title: "Task 1", UserID: 1, IsCompleted: true}
		task2 := &Task{ID: 2, Title: "Task 2", UserID: 1, IsCompleted: true}
		repo.On("Tasks", user.ID, "true").Return([]*Task{task1, task2})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks?taskCompleted=true", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTaskList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task 1")
		assert.Contains(t, w.Body.String(), "Task 2")
	})

	t.Run("RenderTaskListWithIncompleteTasks", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task1 := &Task{ID: 1, Title: "Task 1", UserID: 1, IsCompleted: false}
		task2 := &Task{ID: 2, Title: "Task 2", UserID: 1, IsCompleted: false}
		repo.On("Tasks", user.ID, "false").Return([]*Task{task1, task2})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks?taskCompleted=false", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleTaskList(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task 1")
		assert.Contains(t, w.Body.String(), "Task 2")
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_HandleUpdateSortOrder
func TestDashboardHandlers_HandleUpdateSortOrder(t *testing.T) {
	SetAppDir()

	t.Run("AccessDeniedIfUserNotLoggedIn", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/update-sort-order", nil)

		handler.HandleUpdateSortOrder(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})

	t.Run("InvalidRequestIfJsonDecodeFails", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		formData := `invalid json`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/update-sort-order", strings.NewReader(formData))

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.HandleUpdateSortOrder(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Invalid request")
	})

	t.Run("ErrorUpdatingTaskOrder", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		order := `[{"id": 1, "sortOrder": 2}]`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/update-sort-order", strings.NewReader(order))

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("UpdateTaskSortOrder", 1, user.ID, 2).Return(fmt.Errorf("db error"))

		handler.HandleUpdateSortOrder(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Error updating task order")
	})

	t.Run("UpdateSortOrderSuccess", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		order := `[{"id": 1, "sortOrder": 2}]`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/update-sort-order", strings.NewReader(order))

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("UpdateTaskSortOrder", 1, user.ID, 2).Return(nil)

		handler.HandleUpdateSortOrder(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.JSONEq(t, `{"status": "success"}`, w.Body.String())
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardHandlers_getUserAndTask
func TestDashboardHandlers_getUserAndTask(t *testing.T) {
	SetAppDir()

	t.Run("UserNotLoggedIn", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		user, task := handler.getUserAndTask(w, r)

		assert.Nil(t, user)
		assert.Nil(t, task)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "You need to be logged in to access this feature.")
	})

	t.Run("InvalidTaskID", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/invalid", nil)
		r.SetPathValue("id", "invalid")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.getUserAndTask(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Invalid task ID")
	})

	t.Run("TaskNotFound", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
		r.SetPathValue("id", "999")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		repo.On("TaskByID", 999).Return(nil)

		handler.getUserAndTask(w, r)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Task not found")
	})

	t.Run("AccessDeniedForOtherUser", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{ID: 1, UserID: 2}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		handler.getUserAndTask(w, r)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), "Access denied")
	})

	t.Run("ValidUserAndTask", func(t *testing.T) {
		repo := new(MockDashboardRepository)
		handler := NewDashboardHandler(repo)

		user := &users.User{ID: 1}
		task := &Task{ID: 1, UserID: 1}
		repo.On("TaskByID", 1).Return(task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.SetPathValue("id", "1")

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		userResult, taskResult := handler.getUserAndTask(w, r)

		assert.NotNil(t, userResult)
		assert.NotNil(t, taskResult)
		assert.Equal(t, user.ID, userResult.ID)
		assert.Equal(t, task.ID, taskResult.ID)
	})
}
