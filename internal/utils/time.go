package utils

import (
	"fmt"
	"log/slog"
	"time"
)

// For time.Now() = 2000-01-01 01:00:00.000000000 +0000 UTC
// and timezone = Europe/Moscow
// will return 2000-01-01 04:00:00.000000000 +0000 UTC
func NowWithTimezone(timezone string) (time.Time, error) {

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
		slog.Warn("NowWithTimezone LoadLocation", "timezone", timezone, "err", err)
	}
	now := time.Now().In(loc)
	_, offset := now.Zone()
	now = now.Add(time.Duration(offset) * time.Second).In(time.UTC)
	return now, err
}

// Time can be nil, so *time.Time
func EffectiveTime(time *time.Time, timezone string) (effectiveTime *time.Time) {
	if time == nil {
		now, _ := NowWithTimezone(timezone)
		effectiveTime = &now
	} else {
		effectiveTime = time
	}
	return
}

// Example:
// {{ formatTimeRange .TimeStart .TimeEnd }}
func FormatTimeRange(timeStart time.Time, timeEnd *time.Time, timezone string) string {
	const timeFormat = "15:04"
	const dateTimeFormat = "02 Jan 2006 15:04"

	effectiveEnd := EffectiveTime(timeEnd, timezone)
	duration := effectiveEnd.Sub(timeStart)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	var timeRange string

	if timeEnd == nil {
		now := time.Now()
		isToday := timeStart.Year() == now.Year() && timeStart.YearDay() == now.YearDay()
		if isToday {
			timeRange = fmt.Sprintf("%s - in progress", timeStart.Format(timeFormat))
		} else {
			timeRange = fmt.Sprintf("%s - in progress", timeStart.Format(dateTimeFormat))
		}
	} else {
		if timeStart.Truncate(24 * time.Hour).Equal(timeEnd.Truncate(24 * time.Hour)) {
			// Same date, only time shown
			timeRange = fmt.Sprintf("%s - %s", timeStart.Format(timeFormat), timeEnd.Format(timeFormat))
		} else {
			// Different dates, show date and time
			timeRange = fmt.Sprintf("%s - %s", timeStart.Format(dateTimeFormat), timeEnd.Format(dateTimeFormat))
		}
	}

	if hours > 0 || minutes > 0 {
		timeRange += fmt.Sprintf(" (%dh %dm)", hours, minutes)
	} else {
		timeRange += " (<1m)"
	}

	return timeRange
}
