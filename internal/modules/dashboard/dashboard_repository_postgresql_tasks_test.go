//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardRepositoryPostgres.*
package dashboard

import (
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardRepositoryPostgres_Tasks(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Completed", func(t *testing.T) {
		rows := mockPool.NewRows([]string{"id", "user_id", "title", "description", "color", "sort_order", "is_completed"}).
			AddRow(1, 1, "Task 1", "Description 1", "#FF0000", 1, true).
			AddRow(2, 1, "Task 2", "Description 2", "#00FF00", 2, true)

		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE user_id = \\$1 AND is_completed = true ORDER BY is_completed ASC, sort_order ASC$").
			WithArgs(1).
			WillReturnRows(rows)

		tasks := repo.Tasks(1, "completed")

		require.Len(t, tasks, 2)
		assert.Equal(t, "Task 1", tasks[0].Title)
		assert.Equal(t, "Task 2", tasks[1].Title)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("All", func(t *testing.T) {
		rows := mockPool.NewRows([]string{"id", "user_id", "title", "description", "color", "sort_order", "is_completed"}).
			AddRow(1, 1, "Task 1", "Description 1", "#FF0000", 1, true).
			AddRow(2, 1, "Task 2", "Description 2", "#00FF00", 2, false)

		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE user_id = \\$1 ORDER BY is_completed ASC, sort_order ASC$").
			WithArgs(1).
			WillReturnRows(rows)

		tasks := repo.Tasks(1, "all")

		require.Len(t, tasks, 2)
		assert.Equal(t, "Task 1", tasks[0].Title)
		assert.Equal(t, "Task 2", tasks[1].Title)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("NotCompleted", func(t *testing.T) {
		rows := mockPool.NewRows([]string{"id", "user_id", "title", "description", "color", "sort_order", "is_completed"}).
			AddRow(1, 1, "Task 1", "Description 1", "#FF0000", 1, false).
			AddRow(2, 1, "Task 2", "Description 2", "#00FF00", 2, false)

		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE user_id = \\$1 AND is_completed = false ORDER BY is_completed ASC, sort_order ASC$").
			WithArgs(1).
			WillReturnRows(rows)

		tasks := repo.Tasks(1, "")

		require.Len(t, tasks, 2)
		assert.Equal(t, "Task 1", tasks[0].Title)
		assert.Equal(t, "Task 2", tasks[1].Title)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		mockPool.ExpectQuery(".*").
			WithArgs(1).
			WillReturnError(fmt.Errorf("database error"))

		tasks := repo.Tasks(1, "all")

		require.Len(t, tasks, 0)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("CollectRowsError", func(t *testing.T) {
		// Without is_completed
		rows := mockPool.NewRows([]string{"id", "user_id", "title", "description", "color", "sort_order"}).
			AddRow(1, 1, "Task 1", "Description 1", "#FF0000", 1).
			AddRow(2, 1, "Task 2", "Description 2", "#00FF00", 2)

		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE user_id = \\$1 ORDER BY is_completed ASC, sort_order ASC$").
			WithArgs(1).
			WillReturnRows(rows)

		tasks := repo.Tasks(1, "all")

		require.Len(t, tasks, 0)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_TaskByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		rows := mockPool.NewRows([]string{"id", "user_id", "title", "description", "color", "sort_order", "is_completed"}).
			AddRow(1, 1, "Task 1", "Description 1", "#FF0000", 1, true)

		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE id = \\$1$").
			WithArgs(1).
			WillReturnRows(rows)

		task := repo.TaskByID(1)

		require.NotNil(t, task)
		assert.Equal(t, 1, task.ID)
		assert.Equal(t, "Task 1", task.Title)
		assert.Equal(t, "Description 1", task.Description)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		mockPool.ExpectQuery("^SELECT id, user_id, title, description, color, sort_order, is_completed FROM tasks WHERE id = \\$1$").
			WithArgs(1).
			WillReturnError(fmt.Errorf("no rows"))

		task := repo.TaskByID(1)

		require.Nil(t, task)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		mockPool.ExpectQuery(".*").
			WithArgs(1).
			WillReturnError(fmt.Errorf("database query error"))

		task := repo.TaskByID(1)

		require.Nil(t, task)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_CreateTask(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		task := &Task{
			UserID:      1,
			Title:       "Task 1",
			Description: "Description 1",
			Color:       "#FF0000",
			IsCompleted: false,
		}

		mockPool.ExpectQuery(".*MAX\\(sort_order\\).*").
			WithArgs(task.UserID, task.IsCompleted).
			WillReturnRows(mockPool.NewRows([]string{"max"}).AddRow(5))

		mockPool.ExpectQuery(".*INSERT INTO tasks.*").
			WithArgs(task.UserID, task.Title, task.Description, task.Color, 5+1, task.IsCompleted).
			WillReturnRows(mockPool.NewRows([]string{"id"}).AddRow(10))

		newTaskID, err := repo.CreateTask(task)

		require.NoError(t, err)
		assert.Equal(t, 10, newTaskID)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("InsertError", func(t *testing.T) {
		task := &Task{
			UserID:      1,
			Title:       "Task 1",
			Description: "Description 1",
			Color:       "#FF0000",
			IsCompleted: false,
		}

		mockPool.ExpectQuery(".*MAX\\(sort_order\\).*").
			WithArgs(task.UserID, task.IsCompleted).
			WillReturnRows(mockPool.NewRows([]string{"max"}).AddRow(5))

		mockPool.ExpectQuery(".*INSERT INTO tasks.*").
			WithArgs(task.UserID, task.Title, task.Description, task.Color, 5+1, task.IsCompleted).
			WillReturnError(fmt.Errorf("database insert error"))

		newTaskID, err := repo.CreateTask(task)

		require.Error(t, err)
		assert.Equal(t, 0, newTaskID)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_UpdateTask(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Updated Task",
			Description: "Updated Description",
			Color:       "#00FF00",
			IsCompleted: true,
			SortOrder:   2,
		}

		mockPool.ExpectExec(".*UPDATE tasks.*").
			WithArgs(task.Title, task.Description, task.Color, task.IsCompleted, task.SortOrder, task.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateTask(task)

		require.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("UpdateError", func(t *testing.T) {
		task := &Task{
			ID:          1,
			UserID:      1,
			Title:       "Updated Task",
			Description: "Updated Description",
			Color:       "#00FF00",
			IsCompleted: true,
			SortOrder:   2,
		}

		mockPool.ExpectExec(".*UPDATE tasks.*").
			WithArgs(task.Title, task.Description, task.Color, task.IsCompleted, task.SortOrder, task.ID).
			WillReturnError(fmt.Errorf("database update error"))

		err := repo.UpdateTask(task)

		require.Error(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_DeleteTask(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		taskID := 1

		mockPool.ExpectExec(".*DELETE FROM tasks.*").
			WithArgs(taskID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteTask(taskID)

		require.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DeleteError", func(t *testing.T) {
		taskID := 1

		mockPool.ExpectExec(".*DELETE FROM tasks.*").
			WithArgs(taskID).
			WillReturnError(fmt.Errorf("database delete error"))

		err := repo.DeleteTask(taskID)

		require.Error(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_GetMaxSortOrder(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		userID := 1
		isCompleted := false

		mockPool.ExpectQuery(".*COALESCE\\(MAX\\(sort_order\\), 0\\).*").
			WithArgs(userID, isCompleted).
			WillReturnRows(mockPool.NewRows([]string{"max"}).AddRow(5))

		maxSortOrder := repo.GetMaxSortOrder(userID, isCompleted)

		assert.Equal(t, 5, maxSortOrder)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		userID := 1
		isCompleted := false

		mockPool.ExpectQuery(".*COALESCE\\(MAX\\(sort_order\\), 0\\).*").
			WithArgs(userID, isCompleted).
			WillReturnError(fmt.Errorf("database query error"))

		maxSortOrder := repo.GetMaxSortOrder(userID, isCompleted)

		assert.Equal(t, 0, maxSortOrder)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_UpdateTaskSortOrder(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		taskID := 1
		userID := 1
		sortOrder := 5

		mockPool.ExpectExec("UPDATE tasks SET sort_order = \\$1 WHERE id = \\$2 AND user_id = \\$3").
			WithArgs(sortOrder, taskID, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateTaskSortOrder(taskID, userID, sortOrder)

		require.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		taskID := 1
		userID := 1
		sortOrder := 5

		mockPool.ExpectExec("UPDATE tasks SET sort_order = \\$1 WHERE id = \\$2 AND user_id = \\$3").
			WithArgs(sortOrder, taskID, userID).
			WillReturnError(fmt.Errorf("database update error"))

		err := repo.UpdateTaskSortOrder(taskID, userID, sortOrder)

		require.Error(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
