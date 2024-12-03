//go:build unit

package utils

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestLog.*
func TestLogNewLogHandlerDev(t *testing.T) {
	handler := NewLogHandlerDev()

	if handler == nil {
		t.Fatal("NewLogHandlerDev() returned nil")
	}

	// Check that the handler is configured for debug level
	if handler.Handler.Enabled(context.Background(), slog.LevelDebug) == false {
		t.Errorf("Expected debug level to be enabled")
	}
}

func TestLogHandle(t *testing.T) {
	testCases := []struct {
		name           string
		logLevel       slog.Level
		message        string
		attrs          []slog.Attr
		expectedOutput func(string) bool
	}{
		{
			name:     "Debug log with single attribute",
			logLevel: slog.LevelDebug,
			message:  "Test debug message",
			attrs: []slog.Attr{
				{Key: "!BADKEY", Value: slog.AnyValue([]string{"test", "data"})},
			},
			expectedOutput: func(output string) bool {
				return len(output) > 0 &&
					(output[0] == '\x1b' || // Check for ANSI color codes
						strings.Contains(output, "[]string"))
			},
		},
		{
			name:     "Info log with multiple attributes",
			logLevel: slog.LevelInfo,
			message:  "Test info message",
			attrs: []slog.Attr{
				{Key: "key1", Value: slog.StringValue("value1")},
				{Key: "key2", Value: slog.IntValue(42)},
			},
			expectedOutput: func(output string) bool {
				return strings.Contains(output, "Test info message") &&
					strings.Contains(output, "key1") &&
					strings.Contains(output, "key2")
			},
		},
		{
			name:     "Warn log with multiple attributes",
			logLevel: slog.LevelWarn,
			message:  "Test info message",
			attrs: []slog.Attr{
				{Key: "key1", Value: slog.StringValue("value1")},
				{Key: "key2", Value: slog.IntValue(42)},
			},
			expectedOutput: func(output string) bool {
				return strings.Contains(output, "Test info message") &&
					strings.Contains(output, "key1") &&
					strings.Contains(output, "key2")
			},
		},
		{
			name:     "Error log with panic-related message",
			logLevel: slog.LevelError,
			message:  "panic in /app/some/path",
			attrs:    []slog.Attr{},
			expectedOutput: func(output string) bool {
				return strings.Contains(output, "\x1b[31m") && // Red color
					strings.Contains(output, "panic")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			// Create a record to log
			record := slog.NewRecord(
				time.Now(),
				tc.logLevel,
				tc.message,
				0,
			)
			for _, attr := range tc.attrs {
				record.AddAttrs(attr)
			}

			handler := NewLogHandlerDev()
			var buf bytes.Buffer
			handler.log = log.New(&buf, "", 0)

			err := handler.Handle(context.Background(), record)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify output
			output := buf.String()
			if !tc.expectedOutput(output) {
				t.Errorf("Unexpected output for %s: %s", tc.name, output)
			}
		})
	}
}

func TestMarshalIndentError(t *testing.T) {
	// Create a structure that cannot be serialized
	complexStruct := struct {
		Ch chan int
	}{
		Ch: make(chan int),
	}

	// Capture stdout
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	handler := NewLogHandlerDev()
	record := slog.NewRecord(
		time.Now(),
		slog.LevelDebug,
		"Test message",
		0,
	)
	record.AddAttrs(slog.Any("test", complexStruct))

	err := handler.Handle(context.Background(), record)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLogHighlightPanicAndApp(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Normal message",
			expected: "Normal message",
		},
		{
			input:    "panic something went wrong",
			expected: "\x1b[31mpanic something went wrong\x1b[0m",
		},
		{
			input:    "Message from /app/service",
			expected: "\x1b[31mMessage from /app/service\x1b[0m",
		},
		{
			input:    "Message from time-tracker/module",
			expected: "\x1b[31mMessage from time-tracker/module\x1b[0m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := highlightPanicAndApp(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %q\nGot: %q", tc.expected, result)
			}
		})
	}
}
