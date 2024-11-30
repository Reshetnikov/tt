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
