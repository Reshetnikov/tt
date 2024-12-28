package dashboard

import "time"

type Task struct {
	ID          int
	UserID      int
	Title       string
	Description string
	Color       string
	SortOrder   int
	IsCompleted bool
}

type Record struct {
	ID        int
	TaskID    int
	TimeStart time.Time
	TimeEnd   *time.Time // nullable
	Comment   string

	Task *Task

	StartPercent      float32
	DurationPercent   float32
	Duration          time.Duration
	TimeStartIntraday time.Time
	TimeEndIntraday   time.Time
}

type DailyRecords struct {
	Day     time.Time
	Records []Record
}

type ReportRow struct {
	Task            *Task
	DailyDurations  map[time.Time]time.Duration
	TotalDuration   time.Duration
	DurationPercent float64
}

type ReportData struct {
	ReportRows         []ReportRow
	Days               []time.Time
	DailyTotalDuration map[time.Time]time.Duration
	TotalDuration      time.Duration
}

type DashboardRepository interface {
	Tasks(userID int, taskCompleted string) (tasks []*Task)
	TaskByID(id int) *Task
	CreateTask(task *Task) (int, error)
	UpdateTask(task *Task) error
	DeleteTask(id int) error
	GetMaxSortOrder(userId int, isCompleted bool) (maxSortOrder int)
	UpdateTaskSortOrder(taskID, userID, sortOrder int) error

	RecordsWithTasks(filterRecords FilterRecords) (records []*Record)
	RecordByIDWithTask(recordID int) *Record
	CreateRecord(record *Record) (int, error)
	UpdateRecord(record *Record) error
	DeleteRecord(recordID int) error
	DailyRecords(filterRecords FilterRecords, nowWithTimezone time.Time) (dailyRecords []DailyRecords)
	Reports(
		userID int,
		startInterval time.Time,
		endInterval time.Time,
		nowWithTimezone time.Time,
	) ReportData
}
