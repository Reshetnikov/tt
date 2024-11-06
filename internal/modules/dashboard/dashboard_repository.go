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
	TimeEnd   time.Time
	Comment   string
	Task      *Task
}

type DailyRecords struct {
	Day     time.Time
	Records []Record
}
