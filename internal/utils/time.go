package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

// "2024-W03"
func GetWeekInterval(isoWeek string) (time.Time, time.Time, error) {
	parts := strings.Split(isoWeek, "-W")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, errors.New("invalid ISO week format: " + isoWeek)
	}
	var year, week int
	_, err := fmt.Sscanf(parts[0], "%d", &year)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid year in ISO week format: " + isoWeek)
	}
	_, err = fmt.Sscanf(parts[1], "%d", &week)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid week number in ISO week format: " + isoWeek)
	}
	if week < 1 || week > 53 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid week number: %d", week)
	}

	// Set the date to January 1 of the specified year
	firstJan := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// Calculate the day of the week (ISO 8601 considers Monday as the first day of the week)
	// In Go, the week starts with Sunday (0), so we translate.
	isoWeekday := int(firstJan.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}

	// Shift the first day of the year to the nearest Monday (ISO 8601)
	startOfWeek := firstJan.AddDate(0, 0, -isoWeekday+1)

	// Adding weeks
	startInterval := startOfWeek.AddDate(0, 0, (week-1)*7)
	endInterval := startInterval.AddDate(0, 0, 7).Add(-time.Nanosecond)

	return startInterval, endInterval, nil
}

func GetDateInterval(date time.Time) (time.Time, time.Time) {
	// Calculate the day of the week (ISO 8601: Monday = 1)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday -> 7
	}

	// Calculate the beginning of the week (Monday)
	startInterval := date.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)

	// End of the week (Sunday 23:59:59.999999999)
	endInterval := startInterval.AddDate(0, 0, 7).Add(-time.Nanosecond)

	return startInterval, endInterval
}
