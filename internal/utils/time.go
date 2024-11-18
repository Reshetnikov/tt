package utils

import (
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
