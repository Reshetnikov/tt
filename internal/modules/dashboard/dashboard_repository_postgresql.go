package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
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
func (r *DashboardRepositoryPostgres) Tasks(userID int) (tasks []Task) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, user_id, title, description, color, sort_order, is_completed
		FROM tasks WHERE user_id = $1
		ORDER BY sort_order ASC
	`, userID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres Tasks Query", "err", err)
		return
	}
	defer rows.Close()

	tasks, err = pgx.CollectRows(rows, pgx.RowToStructByName[Task])
	if err != nil {
		slog.Error("DashboardRepositoryPostgres Tasks CollectRows", "err", err)
		return
	}
	return
}

func (r *DashboardRepositoryPostgres) Records(userID int) (records []Record) {
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
		records = append(records, record)
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
