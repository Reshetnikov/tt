//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/dashboard --tags=unit -cover -run TestDashboardRepositoryPostgres.*
package dashboard

import (
	"fmt"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardRepositoryPostgres_RecordsWithTasks(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		filter := FilterRecords{
			UserID:        1,
			RecordID:      123,
			NotRecordID:   125,
			StartInterval: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			InProgress:    true,
		}

		// Diferent tasks for different records
		rows := mockPool.NewRows([]string{
			"id", "task_id", "time_start", "time_end", "comment",
			"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
		}).
			AddRow(123, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), nil, "Comment 1", 1, 1, "Task 1", "Description 1", "#FF0000", 1, false).
			AddRow(124, 1, time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC), nil, "Comment 2", 2, 1, "Task 2", "Description 2", "#FF0000", 1, false)

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.UserID, filter.RecordID, filter.NotRecordID, filter.StartInterval, filter.EndInterval).
			WillReturnRows(rows)

		records := repo.RecordsWithTasks(filter)

		// Diferent tasks for different records
		require.Len(t, records, 2)
		assert.Equal(t, 123, records[0].ID)
		assert.Equal(t, 1, records[0].Task.ID)
		assert.Equal(t, "Task 1", records[0].Task.Title)
		assert.Equal(t, 1, records[0].Task.UserID)
		assert.Equal(t, "Comment 1", records[0].Comment)
		assert.Equal(t, 124, records[1].ID)
		assert.Equal(t, 2, records[1].Task.ID)
		assert.Equal(t, "Task 2", records[1].Task.Title)
		assert.Equal(t, 1, records[1].Task.UserID)
		assert.Equal(t, "Comment 2", records[1].Comment)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("ScanError", func(t *testing.T) {
		filter := FilterRecords{
			UserID:        1,
			RecordID:      123,
			StartInterval: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}

		// "false" instead of false
		// Destination kind 'bool' not supported for value kind 'string' of column 'is_completed'
		rows := mockPool.NewRows([]string{
			"id", "task_id", "time_start", "time_end", "comment",
			"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
		}).
			AddRow(123, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), nil, "Comment 1", 1, 1, "Task 1", "Description 1", "#FF0000", 1, "false")

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.UserID, filter.RecordID, filter.StartInterval, filter.EndInterval).
			WillReturnRows(rows)

		records := repo.RecordsWithTasks(filter)

		require.Len(t, records, 0)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("ExcludeInProgress", func(t *testing.T) {
		filter := FilterRecords{
			UserID:            1,
			ExcludeInProgress: true,
			StartInterval:     time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:       time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}

		timeDnd := time.Date(2024, 12, 2, 10, 0, 0, 0, time.UTC)
		// One task for two records
		rows := mockPool.NewRows([]string{
			"id", "task_id", "time_start", "time_end", "comment",
			"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
		}).
			AddRow(123, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), nil, "Comment 1", 1, 1, "Task 1", "Description 1", "#FF0000", 1, false).
			AddRow(124, 1, time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC), &timeDnd, "Comment 2", 1, 1, "Task 1", "Description 1", "#FF0000", 1, false)

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.UserID, filter.StartInterval, filter.EndInterval).
			WillReturnRows(rows)

		records := repo.RecordsWithTasks(filter)

		// One task for two records
		require.Len(t, records, 2)
		assert.Equal(t, 123, records[0].ID)
		assert.Equal(t, 1, records[0].Task.ID)
		assert.Equal(t, "Task 1", records[0].Task.Title)
		assert.Equal(t, 1, records[0].Task.UserID)
		assert.Equal(t, "Comment 1", records[0].Comment)
		assert.Equal(t, 124, records[1].ID)
		assert.Equal(t, 1, records[1].Task.ID)
		assert.Equal(t, "Task 1", records[1].Task.Title)
		assert.Equal(t, 1, records[1].Task.UserID)
		assert.Equal(t, "Comment 2", records[1].Comment)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		filter := FilterRecords{
			UserID:        1,
			RecordID:      123,
			StartInterval: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.UserID, filter.RecordID, filter.StartInterval, filter.EndInterval).
			WillReturnError(fmt.Errorf("query error"))

		records := repo.RecordsWithTasks(filter)

		require.Len(t, records, 0)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_RecordByIDWithTask(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		recordID := 123
		filter := FilterRecords{RecordID: recordID}

		rows := mockPool.NewRows([]string{
			"id", "task_id", "time_start", "time_end", "comment",
			"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
		}).
			AddRow(123, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), nil, "Comment 1", 1, 1, "Task 1", "Description 1", "#FF0000", 1, false)

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.RecordID).
			WillReturnRows(rows)

		record := repo.RecordByIDWithTask(recordID)

		require.NotNil(t, record)
		assert.Equal(t, recordID, record.ID)
		assert.Equal(t, "Task 1", record.Task.Title)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		recordID := 999
		filter := FilterRecords{RecordID: recordID}

		mockPool.ExpectQuery("SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed").
			WithArgs(filter.RecordID).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}))

		record := repo.RecordByIDWithTask(recordID)

		require.Nil(t, record)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_CreateRecord(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		timeEnd := time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC)
		record := &Record{
			TaskID:    1,
			TimeStart: time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC),
			TimeEnd:   &timeEnd,
			Comment:   "Test Comment",
		}

		mockPool.ExpectQuery(`^INSERT INTO records \(.+\) VALUES \(.+\) RETURNING id`).
			WithArgs(record.TaskID, record.TimeStart, record.TimeEnd, record.Comment).
			WillReturnRows(mockPool.NewRows([]string{"id"}).AddRow(123))

		newRecordID, err := repo.CreateRecord(record)

		require.NoError(t, err)
		assert.Equal(t, 123, newRecordID)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("InsertError", func(t *testing.T) {
		timeEnd := time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC)
		record := &Record{
			TaskID:    1,
			TimeStart: time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC),
			TimeEnd:   &timeEnd,
			Comment:   "Test Comment",
		}

		mockPool.ExpectQuery(`^INSERT INTO records \(.+\) VALUES \(.+\) RETURNING id`).
			WithArgs(record.TaskID, record.TimeStart, record.TimeEnd, record.Comment).
			WillReturnError(fmt.Errorf("database insert error"))

		newRecordID, err := repo.CreateRecord(record)

		require.Error(t, err)
		assert.Equal(t, 0, newRecordID)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_UpdateRecord(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		timeStart := time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC)
		timeEnd := time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC)

		record := &Record{
			ID:        1,
			TaskID:    1,
			TimeStart: timeStart,
			TimeEnd:   &timeEnd,
			Comment:   "Updated Comment",
		}

		mockPool.ExpectExec(`^UPDATE records SET task_id = \$1, time_start = \$2, time_end = \$3, comment = \$4 WHERE id = \$5`).
			WithArgs(record.TaskID, record.TimeStart, record.TimeEnd, record.Comment, record.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateRecord(record)

		require.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("UpdateError", func(t *testing.T) {
		timeStart := time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC)
		timeEnd := time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC)

		record := &Record{
			ID:        1,
			TaskID:    1,
			TimeStart: timeStart,
			TimeEnd:   &timeEnd,
			Comment:   "Updated Comment",
		}

		mockPool.ExpectExec(`^UPDATE records SET task_id = \$1, time_start = \$2, time_end = \$3, comment = \$4 WHERE id = \$5`).
			WithArgs(record.TaskID, record.TimeStart, record.TimeEnd, record.Comment, record.ID).
			WillReturnError(fmt.Errorf("database update error"))

		err := repo.UpdateRecord(record)

		require.Error(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_DeleteRecord(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		recordID := 1

		mockPool.ExpectExec(`^DELETE FROM records WHERE id = \$1`).
			WithArgs(recordID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteRecord(recordID)

		require.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DeleteError", func(t *testing.T) {
		recordID := 1

		mockPool.ExpectExec(`^DELETE FROM records WHERE id = \$1`).
			WithArgs(recordID).
			WillReturnError(fmt.Errorf("database delete error"))

		err := repo.DeleteRecord(recordID)

		require.Error(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_DailyRecords(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		filter := FilterRecords{
			UserID:        1,
			StartInterval: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:   time.Date(2024, 12, 3, 23, 59, 59, 0, time.UTC),
		}

		nowWithTimezone := time.Date(2024, 12, 3, 12, 0, 0, 0, time.UTC)

		mockPool.ExpectQuery(`SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed`).
			WithArgs(filter.UserID, filter.StartInterval, filter.EndInterval).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}).
				AddRow(1, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), &nowWithTimezone, "Test Record 1",
					1, 1, "Task 1", "Description", "#FF0000", 1, false).
				AddRow(2, 2, time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC), nil, "Test Record 2",
					2, 2, "Task 2", "Another Description", "#00FF00", 2, true))

		dailyRecords := repo.DailyRecords(filter, nowWithTimezone)

		require.Len(t, dailyRecords, 3)

		assert.Equal(t, dailyRecords[0].Day, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[0].Records), 1)

		assert.Equal(t, dailyRecords[1].Day, time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[1].Records), 2)

		assert.Equal(t, dailyRecords[2].Day, time.Date(2024, 12, 3, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[2].Records), 2)

		assert.Equal(t, dailyRecords[2].Records[0].Comment, "Test Record 1")
		assert.Equal(t, dailyRecords[2].Records[1].Comment, "Test Record 2")

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("NoRecords", func(t *testing.T) {
		filter := FilterRecords{
			UserID:        1,
			StartInterval: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			EndInterval:   time.Date(2024, 12, 3, 23, 59, 59, 0, time.UTC),
		}

		nowWithTimezone := time.Date(2024, 12, 3, 12, 0, 0, 0, time.UTC)

		mockPool.ExpectQuery(`SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed`).
			WithArgs(filter.UserID, filter.StartInterval, filter.EndInterval).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}))

		dailyRecords := repo.DailyRecords(filter, nowWithTimezone)

		require.Len(t, dailyRecords, 3)

		assert.Equal(t, dailyRecords[0].Day, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[0].Records), 0)

		assert.Equal(t, dailyRecords[1].Day, time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[1].Records), 0)

		assert.Equal(t, dailyRecords[2].Day, time.Date(2024, 12, 3, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, len(dailyRecords[2].Records), 0)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestDashboardRepositoryPostgres_Reports(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewDashboardRepositoryPostgres(mockPool)

	t.Run("Success", func(t *testing.T) {
		startInterval := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		endInterval := time.Date(2024, 12, 3, 23, 59, 59, 0, time.UTC)
		nowWithTimezone := time.Date(2024, 12, 3, 12, 0, 0, 0, time.UTC)
		userID := 1

		mockPool.ExpectQuery(`SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed`).
			WithArgs(userID, startInterval, endInterval).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}).
				AddRow(1, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), &nowWithTimezone, "Comment 1",
					1, userID, "Task 1", "Description 1", "#FF0000", 1, false).
				AddRow(2, 2, time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC), nil, "Comment 2",
					2, userID, "Task 2", "Description 2", "#00FF00", 2, true))

		report := repo.Reports(userID, startInterval, endInterval, nowWithTimezone)

		require.Len(t, report.ReportRows, 2)
		assert.Equal(t, "Task 1", report.ReportRows[0].Task.Title)
		assert.Equal(t, "Task 2", report.ReportRows[1].Task.Title)

		assert.Len(t, report.Days, 3)
		assert.Equal(t, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), report.Days[0])
		assert.Equal(t, time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC), report.Days[1])
		assert.Equal(t, time.Date(2024, 12, 3, 0, 0, 0, 0, time.UTC), report.Days[2])

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("SortOrder", func(t *testing.T) {
		startInterval := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		endInterval := time.Date(2024, 12, 3, 23, 59, 59, 0, time.UTC)
		nowWithTimezone := time.Date(2024, 12, 3, 12, 0, 0, 0, time.UTC)
		userID := 1

		mockPool.ExpectQuery(`SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed`).
			WithArgs(userID, startInterval, endInterval).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}).
				AddRow(1, 1, time.Date(2024, 12, 1, 8, 0, 0, 0, time.UTC), &nowWithTimezone, "Comment 1",
					1, userID, "Task A", "Description A", "#FF0000", 2, false).
				AddRow(2, 2, time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC), nil, "Comment 2",
					2, userID, "Task B", "Description B", "#00FF00", 1, false).
				AddRow(3, 3, time.Date(2024, 12, 1, 10, 0, 0, 0, time.UTC), nil, "Comment 3",
					3, userID, "Task C", "Description C", "#0000FF", 3, false))

		report := repo.Reports(userID, startInterval, endInterval, nowWithTimezone)

		require.Len(t, report.ReportRows, 3)
		assert.Equal(t, "Task B", report.ReportRows[0].Task.Title) // SortOrder = 1
		assert.Equal(t, "Task A", report.ReportRows[1].Task.Title) // SortOrder = 2
		assert.Equal(t, "Task C", report.ReportRows[2].Task.Title) // SortOrder = 3

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("NoRecords", func(t *testing.T) {
		startInterval := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		endInterval := time.Date(2024, 12, 3, 23, 59, 59, 0, time.UTC)
		nowWithTimezone := time.Date(2024, 12, 3, 12, 0, 0, 0, time.UTC)
		userID := 1

		mockPool.ExpectQuery(`SELECT r.id, r.task_id, r.time_start, r.time_end, r.comment, t.id, t.user_id, t.title, t.description, t.color, t.sort_order, t.is_completed`).
			WithArgs(userID, startInterval, endInterval).
			WillReturnRows(mockPool.NewRows([]string{
				"id", "task_id", "time_start", "time_end", "comment",
				"id", "user_id", "title", "description", "color", "sort_order", "is_completed",
			}))

		report := repo.Reports(userID, startInterval, endInterval, nowWithTimezone)

		require.Len(t, report.ReportRows, 0)
		assert.Len(t, report.Days, 3)
		assert.Equal(t, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), report.Days[0])
		assert.Equal(t, time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC), report.Days[1])
		assert.Equal(t, time.Date(2024, 12, 3, 0, 0, 0, 0, time.UTC), report.Days[2])

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
