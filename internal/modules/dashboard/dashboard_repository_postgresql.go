package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardRepositoryPostgres отвечает за доступ к данным
type DashboardRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewDashboardRepositoryPostgres создает новый репозиторий для работы с базой данных
func NewDashboardRepositoryPostgres(db *pgxpool.Pool) *DashboardRepositoryPostgres {
	return &DashboardRepositoryPostgres{db: db}
}

// FetchTasks извлекает задачи пользователя
func (r *DashboardRepositoryPostgres) FetchTasks(userID int) (tasks []Task) {
	rows, err := r.db.Query(context.Background(), "SELECT id, user_id, title, description, color, is_completed FROM tasks WHERE user_id = $1", userID)
	// slog.Debug("FetchTasks", "rows", rows, "user_id", userID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres FetchTasks Query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Color, &task.IsCompleted); err != nil {
			slog.Error("DashboardRepositoryPostgres FetchTasks Scan", "err", err)
			return
		}
		tasks = append(tasks, task)
	}
	return
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
