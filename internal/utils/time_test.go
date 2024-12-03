//go:build unit

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestTime.*
func TestTime_NowWithTimezone(t *testing.T) {
	testCases := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{"Valid Moscow Timezone", "Europe/Moscow", false},
		{"Valid UTC Timezone", "UTC", false},
		{"Invalid Timezone", "Invalid/Timezone", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now, err := NowWithTimezone(tc.timezone)
			if (err != nil) != tc.wantErr {
				t.Errorf("NowWithTimezone() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if now.IsZero() {
				t.Error("NowWithTimezone() returned zero time")
			}
		})
	}
}

func TestTime_EffectiveTime(t *testing.T) {
	timezone := "Europe/Moscow"

	now := time.Now()
	effective := EffectiveTime(&now, timezone)
	assert.Equal(t, &now, effective)

	effective = EffectiveTime(nil, timezone)
	assert.NotNil(t, effective)
}

func TestTime_FormatTimeRange(t *testing.T) {
	timezone := "Europe/Moscow"

	start := time.Date(2024, 12, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 1, 12, 30, 0, 0, time.UTC)

	result := FormatTimeRange(start, &end, timezone)
	assert.Contains(t, result, "10:00 - 12:30")
	assert.Contains(t, result, "2h 30m")

	result = FormatTimeRange(start, &start, timezone)
	assert.Contains(t, result, "10:00 - 10:0")
	assert.Contains(t, result, "(<1m)")

	start = time.Date(2024, 12, 1, 23, 0, 0, 0, time.UTC)
	end = time.Date(2024, 12, 2, 03, 30, 0, 0, time.UTC)
	result = FormatTimeRange(start, &end, timezone)
	assert.Contains(t, result, "01 Dec 2024 23:00 - 02 Dec 2024 03:30")
	assert.Contains(t, result, "4h 30m")

	result = FormatTimeRange(start, nil, timezone)
	assert.Contains(t, result, "in progress")

	result = FormatTimeRange(time.Now(), nil, timezone)
	assert.Contains(t, result, "in progress")
}

func TestTime_FormatDuration(t *testing.T) {
	duration := 2*time.Hour + 15*time.Minute
	assert.Equal(t, "2h 15m", FormatDuration(duration))

	duration = 3 * time.Hour
	assert.Equal(t, "3h", FormatDuration(duration))

	duration = 45 * time.Minute
	assert.Equal(t, "45m", FormatDuration(duration))

	duration = 0
	assert.Equal(t, "-", FormatDuration(duration))
}

func TestTime_GetWeekInterval(t *testing.T) {
	testCases := []struct {
		name              string
		weekStr           string
		isWeekStartMonday bool
		expectErr         bool
	}{
		{"Valid ISO week Monday start", "2023-W03", true, false},
		{"Valid ISO week Monday start", "2024-W03", true, false},
		{"Valid ISO week Sunday start", "2024-W03", false, false},
		{"Invalid format", "2024W03", true, true},
		{"Invalid year", "abcd-W03", true, true},
		{"Invalid week number", "2024-W54", true, true},
		{"Invalid week format", "2024-WXX", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start, end, err := GetWeekInterval(tc.weekStr, tc.isWeekStartMonday)
			if (err != nil) != tc.expectErr {
				t.Errorf("GetWeekInterval() error = %v, expectErr %v", err, tc.expectErr)
				return
			}
			if !tc.expectErr {
				if start.IsZero() || end.IsZero() {
					t.Error("GetWeekInterval() returned zero times")
				}
			}
		})
	}
}

func TestTime_FormatISOWeek(t *testing.T) {
	date := time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC) // Wednesday in week 3
	isoWeek := FormatISOWeek(date, true)
	assert.Equal(t, "2024-W52", isoWeek)

	date = time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC) // Wednesday in week 3
	isoWeek = FormatISOWeek(date, false)
	assert.Equal(t, "2025-W01", isoWeek)

	date = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC) // Wednesday in week 3
	isoWeek = FormatISOWeek(date, true)
	assert.Equal(t, "2022-W52", isoWeek)

	date = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC) // Wednesday in week 3
	isoWeek = FormatISOWeek(date, false)
	assert.Equal(t, "2023-W01", isoWeek)
}

func TestTime_FormatTimeForInput(t *testing.T) {
	date := time.Date(2024, 12, 1, 10, 0, 0, 0, time.UTC)
	result := FormatTimeForInput(&date)
	assert.Equal(t, "2024-12-01T10:00", result)

	result = FormatTimeForInput(nil)
	assert.Equal(t, "", result)
}
