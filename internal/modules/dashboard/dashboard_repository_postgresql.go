package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardRepositoryPostgres struct {
	db *pgxpool.Pool
}

func NewDashboardRepositoryPostgres(db *pgxpool.Pool) *DashboardRepositoryPostgres {
	return &DashboardRepositoryPostgres{db: db}
}

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

func (r *DashboardRepositoryPostgres) Records(userID int) (records []*Record) {
	rows, err := r.db.Query(context.Background(), `
		SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment
		FROM records r
		JOIN tasks t ON r.task_id = t.id
		WHERE t.user_id = $1
		ORDER BY r.time_start ASC;
	`, userID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres Records Query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.TaskID, &record.TimeStart, &record.TimeEnd, &record.Comment)
		if err != nil {
			slog.Error("DashboardRepositoryPostgres Records Scan", "err", err)
			return
		}
		records = append(records, &record)
	}
	return
}

func (r *DashboardRepositoryPostgres) RecordsWithTasks(userID int) (records []*Record) {
	rows, err := r.db.Query(context.Background(), `
		SELECT 
			r.id, r.task_id, r.time_start, r.time_end, r.comment,
			t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed
		FROM records r
		JOIN tasks t ON r.task_id = t.id
		WHERE t.user_id = $1
		ORDER BY r.time_start ASC;
	`, userID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres RecordsWithTasks Query", "err", err)
		return
	}
	defer rows.Close()

	taskMap := make(map[int]*Task)
	for rows.Next() {
		var record Record
		var task Task

		err := rows.Scan(
			&record.ID, &record.TaskID, &record.TimeStart, &record.TimeEnd, &record.Comment,
			&task.ID, &task.UserID, &task.Title, &task.Description, &task.Color, &task.SortOrder, &task.IsCompleted,
		)
		if err != nil {
			slog.Error("DashboardRepositoryPostgres RecordsWithTasks Scan", "err", err)
			return
		}
		if existingTask, exists := taskMap[task.ID]; exists {
			record.Task = existingTask
		} else {
			taskCopy := task
			record.Task = &taskCopy
			taskMap[task.ID] = &taskCopy
		}

		records = append(records, &record)
	}
	return
}

/*func (r *DashboardRepositoryPostgres) RecordsAndTasks(userID int) (records []*Record, tasks []*Task) {
	tasks = r.Tasks(userID, "all")
	if len(tasks) == 0 {
		return
	}

	tasksMap := make(map[int]*Task, len(tasks))
	for _, task := range tasks {
		tasksMap[task.ID] = task
	}

	records = r.Records(userID)
	for _, record := range records {
		if task, exists := tasksMap[record.TaskID]; exists {
			record.Task = task
		} else {
			slog.Error("DashboardRepositoryPostgres RecordsWithTasks Task Not Found", "record", record)
		}
	}
	return
}*/

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

func (r *DashboardRepositoryPostgres) CreateTask(task *Task) (int, error) {
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
		return 0, err
	}
	return newTaskID, nil
}

func (r *DashboardRepositoryPostgres) GetTaskByID(id int) (*Task, error) {
	var task Task
	err := r.db.QueryRow(context.Background(), `
		SELECT id, user_id, title, description, color, sort_order, is_completed
		FROM tasks WHERE id = $1
	`, id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Color, &task.SortOrder, &task.IsCompleted)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres GetTaskByID Query", "err", err)
		return nil, err
	}
	return &task, nil
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

// userID is needed for access control instead of validation
func (repo *DashboardRepositoryPostgres) UpdateTaskSortOrder(taskID, userID, sortOrder int) error {
	query := `UPDATE tasks SET sort_order = $1 WHERE id = $2 AND user_id = $3`
	_, err := repo.db.Exec(context.Background(), query, sortOrder, taskID, userID)
	return err
}

// FetchWeeklyRecords извлекает записи за неделю
func (r *DashboardRepositoryPostgres) FetchWeeklyRecords(userID int, startOfWeek time.Time) (weeklyRecords []DailyRecords) {
	for i := 0; i < 7; i++ {
		day := startOfWeek.AddDate(0, 0, i)
		rows, err := r.db.Query(context.Background(), "SELECT id, task_id, time_start, time_end, comment FROM records WHERE time_start >= $1 AND time_start < $2", day, day.Add(24*time.Hour))
		if err != nil {
			slog.Error("DashboardRepositoryPostgres FetchWeeklyRecords Query", "err", err)
			return
		}
		defer rows.Close()

		var records []Record
		for rows.Next() {
			var record Record
			if err := rows.Scan(&record.ID, &record.TaskID, &record.TimeStart, &record.TimeEnd, &record.Comment); err != nil {
				slog.Error("DashboardRepositoryPostgres FetchWeeklyRecords Scan", "err", err)
				return
			}
			records = append(records, record)
		}

		weeklyRecords = append(weeklyRecords, DailyRecords{
			Day:     day,
			Records: records,
		})
	}

	return
}
