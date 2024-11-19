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
	NotRecordID   int
	StartInterval time.Time
	EndInterval   time.Time
	InProgress    bool
	// If StartInterval is in the future, then time_end IS NULL entries should be excluded.
	// Because we can consider time_end = now() and now() < StartInterval, i.e. time_end < StartInterval.
	ExcludeInProgress bool
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

	// NotRecordID
	if filterRecords.NotRecordID > 0 {
		filters = append(filters, fmt.Sprintf("r.id != $%d", argIndex))
		args = append(args, filterRecords.NotRecordID)
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

	// InProgress
	if filterRecords.InProgress {
		filters = append(filters, "r.time_end IS NULL")
	}

	// ExcludeInProgress
	if filterRecords.InProgress {
		filters = append(filters, "r.time_end IS NOT NULL")
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

func (r *DashboardRepositoryPostgres) CreateRecord(record *Record) (int, error) {
	var newRecordID int
	err := r.db.QueryRow(context.Background(), `
        INSERT INTO records (task_id, time_start, time_end, comment)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, record.TaskID, record.TimeStart, record.TimeEnd, record.Comment).Scan(&newRecordID)
	if err != nil {
		slog.Error("DashboardRepositoryPostgres CreateRecord QueryRow", "err", err)
		return 0, err
	}
	return newRecordID, nil
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

func (r *DashboardRepositoryPostgres) DailyRecords(filterRecords FilterRecords, nowWithTimezone *time.Time) (dailyRecords []DailyRecords) {
	if filterRecords.StartInterval.IsZero() || filterRecords.EndInterval.IsZero() {
		return
	}
	if filterRecords.StartInterval.After(filterRecords.EndInterval) {
		return
	}

	// Приведение интервала к началу суток
	startInterval := filterRecords.StartInterval.Truncate(24 * time.Hour)
	endInterval := filterRecords.EndInterval.Truncate(24 * time.Hour)

	// Мапа для распределения записей по дням
	dayMap := make(map[time.Time][]Record)
	for d := startInterval; !d.After(endInterval); d = d.Add(24 * time.Hour) {
		dayMap[d] = []Record{}
	}

	records := r.RecordsWithTasks(filterRecords)

	// Обработка записей
	for _, record := range records {
		start := record.TimeStart
		end := record.TimeEnd

		// Если TimeEnd == nil, запись распространяется до конца интервала
		if end == nil {
			temp := endInterval.Add(24 * time.Hour)
			end = &temp
		}

		// Распределение записи по дням
		for d := start.Truncate(24 * time.Hour); !d.After(end.Truncate(24 * time.Hour)); d = d.Add(24 * time.Hour) {
			// Проверяем, входит ли день в указанный интервал
			if d.Before(startInterval) || d.After(endInterval) {
				continue
			}

			// Копия записи для конкретного дня
			dailyRecord := *record
			startOfDay := d
			endOfDay := d.Add(24 * time.Hour)

			// Начало записи для текущего дня
			dailyStart := start
			if dailyStart.Before(startOfDay) {
				dailyStart = startOfDay
			}

			// Конец записи для текущего дня
			dailyEnd := *end
			if dailyEnd.After(endOfDay) {
				dailyEnd = endOfDay
			}

			// Вычисление процентов и продолжительности
			totalDaySeconds := float64(24 * time.Hour / time.Second)
			dailyRecord.StartPercent = int(float64(dailyStart.Sub(startOfDay)) / totalDaySeconds * 100)
			dailyRecord.DurationPercent = int(float64(dailyEnd.Sub(dailyStart)) / totalDaySeconds * 100)
			dailyRecord.Duration = dailyEnd.Sub(dailyStart)

			// Добавление записи в день
			dayMap[d] = append(dayMap[d], dailyRecord)
		}
	}

	// Преобразование мапы в массив
	for d := startInterval; !d.After(endInterval); d = d.Add(24 * time.Hour) {
		dailyRecords = append(dailyRecords, DailyRecords{
			Day:     d,
			Records: dayMap[d],
		})
	}

	return dailyRecords
}
