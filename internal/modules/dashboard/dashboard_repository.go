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
	DurationHM          string
	TimeStartIntraday string
	TimeEndIntraday   string
}

type DailyRecords struct {
	Day     time.Time
	Records []Record
}
