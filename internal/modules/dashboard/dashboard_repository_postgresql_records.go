package dashboard

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
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
		filters = append(filters, fmt.Sprintf("(r.time_end > $%d OR r.time_end IS NULL)", argIndex))
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
	if filterRecords.ExcludeInProgress {
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

func (r *DashboardRepositoryPostgres) DailyRecords(filterRecords FilterRecords, nowWithTimezone time.Time) (dailyRecords []DailyRecords) {
	dateFirstDay := filterRecords.StartInterval.Truncate(24 * time.Hour)
	dateLastDay := filterRecords.EndInterval.Truncate(24 * time.Hour)

	dayMap := make(map[time.Time][]Record)
	for d := dateFirstDay; !d.After(dateLastDay); d = d.Add(24 * time.Hour) {
		dayMap[d] = []Record{}
	}

	records := r.RecordsWithTasks(filterRecords)

	for _, record := range records {
		timeEnd := record.TimeEnd
		if timeEnd == nil {
			timeEnd = &nowWithTimezone
		}
		lastDayRecord := timeEnd.Truncate(24 * time.Hour)

		dayRecord := record.TimeStart.Truncate(24 * time.Hour)
		for ; !dayRecord.After(lastDayRecord) && !dayRecord.Equal(*timeEnd); dayRecord = dayRecord.Add(24 * time.Hour) {
			// Because there will be different StartPercent, DurationPercent, Duration if the recording lasts several days
			recordCopy := *record
			dayEndRecord := dayRecord.Add(24 * time.Hour)

			timeStartIntraday := recordCopy.TimeStart
			if timeStartIntraday.Before(dayRecord) {
				timeStartIntraday = dayRecord
			}

			timeEndIntraday := *timeEnd
			// D("t", "timeEndIntraday", timeEndIntraday, "dayEndRecord", dayEndRecord, "1", timeEndIntraday.After(dayEndRecord))
			if timeEndIntraday.After(dayEndRecord) {
				timeEndIntraday = dayEndRecord
			}

			recordCopy.TimeStartIntraday = timeStartIntraday
			recordCopy.TimeEndIntraday = timeEndIntraday
			recordCopy.Duration = timeEndIntraday.Sub(timeStartIntraday)

			totalDaySeconds := float32(86400)
			recordCopy.StartPercent = float32(timeStartIntraday.Sub(dayRecord)/time.Second) / totalDaySeconds * 100
			recordCopy.DurationPercent = float32(timeEndIntraday.Sub(timeStartIntraday)/time.Second) / totalDaySeconds * 100

			dayMap[dayRecord] = append(dayMap[dayRecord], recordCopy)
		}
	}

	for d := dateFirstDay; !d.After(dateLastDay); d = d.Add(24 * time.Hour) {
		dailyRecords = append(dailyRecords, DailyRecords{
			Day:     d,
			Records: dayMap[d],
		})
	}

	return dailyRecords
}

func (r *DashboardRepositoryPostgres) Reports(
	userID int,
	startInterval time.Time,
	endInterval time.Time,
	nowWithTimezone time.Time,
) ReportData {
	recordsFilter := FilterRecords{
		UserID:        userID,
		StartInterval: startInterval,
		EndInterval:   endInterval,
	}
	dailyRecords := r.DailyRecords(recordsFilter, nowWithTimezone)

	var reportRows []ReportRow
	var days []time.Time
	var totalDuration time.Duration
	reportRowsMap := make(map[int]*ReportRow)
	dailyTotalDuration := make(map[time.Time]time.Duration)

	for _, dailyRecord := range dailyRecords {
		days = append(days, dailyRecord.Day)
		for _, record := range dailyRecord.Records {
			// Create reportRowsMap[record.TaskID]
			if _, exists := reportRowsMap[record.TaskID]; !exists {
				reportRowsMap[record.TaskID] = &ReportRow{
					Task:           record.Task,
					DailyDurations: make(map[time.Time]time.Duration),
				}
			}
			// Filling reportRowsMap[record.TaskID]
			reportRowsMap[record.TaskID].DailyDurations[dailyRecord.Day] += record.Duration
			reportRowsMap[record.TaskID].TotalDuration += record.Duration
			dailyTotalDuration[dailyRecord.Day] += record.Duration
			totalDuration += record.Duration
		}
	}
	// Calculate DurationPercent
	for _, row := range reportRowsMap {
		if totalDuration > 0 {
			row.DurationPercent = float64(row.TotalDuration) / float64(totalDuration) * 100
		}
	}
	// reportRowsMap -> reportRows slice
	for _, row := range reportRowsMap {
		reportRows = append(reportRows, *row)
	}
	// Sort reportRows
	sort.Slice(reportRows, func(i, j int) bool {
		if reportRows[i].Task.IsCompleted != reportRows[j].Task.IsCompleted {
			return !reportRows[i].Task.IsCompleted
		}
		return reportRows[i].Task.SortOrder < reportRows[j].Task.SortOrder
	})

	return ReportData{
		ReportRows:         reportRows,
		Days:               days,
		DailyTotalDuration: dailyTotalDuration,
		TotalDuration:      totalDuration,
	}
}
