package dashboard

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func (r *DashboardRepositoryPostgres) Tasks(userID int, taskCompleted string) (tasks []*Task) {
	query := `
		SELECT id, user_id, title, description, color, sort_order, is_completed
		FROM tasks WHERE user_id = $1
	`
	switch taskCompleted {
	case "completed":
		query += " AND is_completed = true"
	case "all":
		// We do not add any conditions for "all"
	default:
		query += " AND is_completed = false"
	}
	query += " ORDER BY is_completed ASC, sort_order ASC"
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres Tasks Query", "err", err)
		return
	}
	defer rows.Close()

	taskValues, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
	if err != nil {
		slog.Error("DashboardRepositoryPostgres Tasks CollectRows", "err", err)
		return
	}

	// []Task to []*Task
	for _, t := range taskValues {
		task := t
		tasks = append(tasks, &task)
	}
	return
}

func (r *DashboardRepositoryPostgres) TaskByID(id int) *Task {
	var task Task
	err := r.db.QueryRow(context.Background(), `
		SELECT id, user_id, title, description, color, sort_order, is_completed
		FROM tasks WHERE id = $1
	`, id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Color, &task.SortOrder, &task.IsCompleted)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres TaskByID Query", "err", err)
		return nil
	}
	return &task
}

func (r *DashboardRepositoryPostgres) CreateTask(task *Task) int {
	maxSortOrder := r.GetMaxSortOrder(task.UserID, task.IsCompleted)
	task.SortOrder = maxSortOrder + 1

	var newTaskID int
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO tasks (user_id, title, description, color, sort_order, is_completed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, task.UserID, task.Title, task.Description, task.Color, task.SortOrder, task.IsCompleted).Scan(&newTaskID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres CreateTask QueryRow", "err", err)
		return 0
	}
	return newTaskID
}

func (r *DashboardRepositoryPostgres) UpdateTask(task *Task) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE tasks
		SET title = $1, description = $2, color = $3, is_completed = $4, sort_order = $5
		WHERE id = $6
	`, task.Title, task.Description, task.Color, task.IsCompleted, task.SortOrder, task.ID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres UpdateTask Query", "err", err)
		return err
	}
	return nil
}

func (r *DashboardRepositoryPostgres) DeleteTask(id int) error {
	_, err := r.db.Exec(context.Background(), `
		DELETE FROM tasks WHERE id = $1
	`, id)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres DeleteTask Query", "err", err)
		return err
	}
	return nil
}

func (r *DashboardRepositoryPostgres) GetMaxSortOrder(userId int, isCompleted bool) (maxSortOrder int) {
	err := r.db.QueryRow(context.Background(), `
        SELECT COALESCE(MAX(sort_order), 0) 
        FROM tasks 
        WHERE user_id = $1 AND is_completed = $2
    `, userId, isCompleted).Scan(&maxSortOrder)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres CreateTask QueryRow", "err", err)
	}
	return
}

// userID is needed for access control instead of validation
func (repo *DashboardRepositoryPostgres) UpdateTaskSortOrder(taskID, userID, sortOrder int) error {
	query := `UPDATE tasks SET sort_order = $1 WHERE id = $2 AND user_id = $3`
	_, err := repo.db.Exec(context.Background(), query, sortOrder, taskID, userID)
	return err
}
