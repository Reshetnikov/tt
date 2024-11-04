package utils

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strings"
)

type LogHandlerDev struct {
    slog.Handler
    log *log.Logger
}

func NewLogHandlerDev() *LogHandlerDev {
    opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
    h := &LogHandlerDev{
        Handler: slog.NewJSONHandler(os.Stdout, opts),
        log:     log.New(os.Stdout, "", 0),
    }

    return h
}

func (h *LogHandlerDev) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

    switch r.Level {
    case slog.LevelDebug:
        level = StrColor(level,Colors.Green)
    case slog.LevelInfo:
        level = StrColor(level, Colors.Blue)
    case slog.LevelWarn:
        level = StrColor(level, Colors.Yellow)
    case slog.LevelError:
        level = Colors.Red + level + Colors.Reset
    }

    timeStr := r.Time.Format("[15:05:05.000]")
	msg := highlightPanicAndApp(r.Message)
    h.log.Println(timeStr, level, msg)

    r.Attrs(func(a slog.Attr) bool {
        key := a.Key
        val := a.Value.Any()

        h.log.Printf("%s%v:%s %s%T%s = %s%+v%s\n",
            Colors.Green, key, Colors.Reset,     // Name
            Colors.Blue, val, Colors.Reset,      // Type
            Colors.Yellow, val, Colors.Reset)    // Value

        b, err := json.MarshalIndent(val, "", "  ")
        if err != nil {
            return false
        }
        h.log.Println(string(b))

        return true
    })

    return nil
}

func highlightPanicAndApp(logMessage string) string {
	lines := strings.Split(logMessage, "\n")
	var highlightedLines []string

	for _, line := range lines {
		if strings.Contains(line, "panic ") || strings.Contains(line, "/app/") || strings.Contains(line, "time-tracker/") {
            line = StrColor(line, Colors.Red)
		}
		highlightedLines = append(highlightedLines, line)
	}

	return strings.Join(highlightedLines, "\n")
}