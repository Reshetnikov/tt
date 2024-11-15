package dashboard

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type FilterRecords struct {
	UserID        int
	RecordID      int
	StartInterval time.Time
	EndInterval   time.Time
}

func (r *DashboardRepositoryPostgres) RecordsWithTasks(filterRecords FilterRecords) (records []*Record) {
	query := `
        SELECT 
            r.id, r.task_id, r.time_start, r.time_end, r.comment,
            t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed
        FROM records r
        JOIN tasks t ON r.task_id = t.id
    `
	filters := []string{}
	args := []interface{}{}
	argIndex := 1

	// UserID
	if filterRecords.UserID > 0 {
		filters = append(filters, fmt.Sprintf("t.user_id = $%d", argIndex))
		args = append(args, filterRecords.UserID)
		argIndex++
	}

	// RecordID
	if filterRecords.RecordID > 0 {
		filters = append(filters, fmt.Sprintf("r.id = $%d", argIndex))
		args = append(args, filterRecords.RecordID)
		argIndex++
	}

	// Recording must end after interval starts or does not end
	if !filterRecords.StartInterval.IsZero() {
		filters = append(filters, fmt.Sprintf("(r.time_end >= $%d OR r.time_end IS NULL)", argIndex))
		args = append(args, filterRecords.StartInterval)
		argIndex++
	}

	// Recording must start before the end of the interval
	if !filterRecords.EndInterval.IsZero() {
		filters = append(filters, fmt.Sprintf("r.time_start < $%d", argIndex))
		args = append(args, filterRecords.EndInterval)
		argIndex++
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY r.time_start ASC"

	rows, err := r.db.Query(context.Background(), query, args...)
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

func (r *DashboardRepositoryPostgres) RecordByIDWithTask(recordID int) *Record {
	records := r.RecordsWithTasks(FilterRecords{
		RecordID: recordID,
	})
	if len(records) == 0 {
		return nil
	}
	return records[0]
}

func (r *DashboardRepositoryPostgres) CreateRecord(record *Record) int {
	var newRecordID int
	err := r.db.QueryRow(context.Background(), `
        INSERT INTO records (task_id, time_start, time_end, comment)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, record.TaskID, record.TimeStart, record.TimeEnd, record.Comment).Scan(&newRecordID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres CreateRecord QueryRow", "err", err)
		return 0
	}
	return newRecordID
}

func (r *DashboardRepositoryPostgres) UpdateRecord(record *Record) error {
	_, err := r.db.Exec(context.Background(), `
        UPDATE records
        SET task_id = $1, time_start = $2, time_end = $3, comment = $4
        WHERE id = $5
    `, record.TaskID, record.TimeStart, record.TimeEnd, record.Comment, record.ID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres UpdateRecord Query", "err", err)
		return err
	}
	return nil
}

func (r *DashboardRepositoryPostgres) DeleteRecord(recordID int) error {
	_, err := r.db.Exec(context.Background(), `
        DELETE FROM records WHERE id = $1
    `, recordID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres DeleteRecord Query", "err", err)
		return err
	}
	return nil
}

// FetchWeeklyRecords retrieves records for a week
/*func (r *DashboardRepositoryPostgres) FetchWeeklyRecords(userID int, startOfWeek time.Time) (weeklyRecords []DailyRecords) {
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
}*/

/*func (r *DashboardRepositoryPostgres) Records(userID int) (records []*Record) {
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
}*/
