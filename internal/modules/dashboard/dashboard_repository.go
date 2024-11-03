package dashboard

import "time"

type Task struct {
    ID          int
    UserID     int
    Title       string
    Description string
    Color       string
    IsCompleted bool
}

// Record представляет запись о времени
type Record struct {
    ID          int
    TaskID     int
    TimeStart   time.Time
    TimeEnd     time.Time
    Comment     string
}

// DailyRecords представляет записи, сгруппированные по дням
type DailyRecords struct {
    Day     time.Time
    Records []Record
}

// DashboardData содержит данные, необходимые для шаблона dashboard
type DashboardData struct {
    Tasks         []Task
    WeeklyRecords []DailyRecords
    SelectedWeek  time.Time
}