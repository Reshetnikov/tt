package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

type LogHandlerDev struct {
    slog.Handler
    l *log.Logger
}

func (h *LogHandlerDev) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

    switch r.Level {
    case slog.LevelDebug:
        level = Colors.Green + level + Colors.Reset
    case slog.LevelInfo:
        level = Colors.Blue + level + Colors.Reset
    case slog.LevelWarn:
        level = Colors.Yellow + level + Colors.Reset
    case slog.LevelError:
        level = Colors.Red + level + Colors.Reset
    }

    fields := make(map[string]interface{}, r.NumAttrs())
    r.Attrs(func(a slog.Attr) bool {
        fields[a.Key] = a.Value.Any()

        return true
    })

    b, err := json.MarshalIndent(fields, "", "  ")
    if err != nil {
        return err
    }

    timeStr := r.Time.Format("[15:05:05.000]")
	msg := highlightPanicAndApp(r.Message)
    h.l.Println(timeStr, level, msg, string(b))

    return nil
}

func NewLogHandlerDev() *LogHandlerDev {
    h := &LogHandlerDev{
        Handler: slog.NewJSONHandler(os.Stdout, nil),
        l:       log.New(os.Stdout, "", 0),
    }

    return h
}

func highlightPanicAndApp(logMessage string) string {
	lines := strings.Split(logMessage, "\n")
	var highlightedLines []string

	for _, line := range lines {
		// Проверяем, содержит ли строка "panic" или "/app/"
		if strings.Contains(line, "panic ") || strings.Contains(line, "/app/") || strings.Contains(line, "time-tracker/") {
			// Подсвечиваем красным
			line = fmt.Sprintf("\033[31m%s\033[0m", line)
		}
		highlightedLines = append(highlightedLines, line)
	}

	return strings.Join(highlightedLines, "\n")
}