package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Replacements(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		// Replace time with relatively concise, yet standard format.
		a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
	case slog.LevelKey:
		// Customize the name of the level key and the output string, including
		// custom level values.
		v := a.Value.Any().(slog.Level)

		// Renaming the log levels based on their priority.
		switch {
		case v <= slog.Level(-8):
			a.Value = slog.StringValue("TRACE")
		case v <= slog.LevelDebug:
			a.Value = slog.StringValue("DEBUG")
		case v <= slog.LevelInfo:
			a.Value = slog.StringValue("INFO")
		case v <= slog.Level(2):
			a.Value = slog.StringValue("NOTICE")
		case v <= slog.LevelWarn:
			a.Value = slog.StringValue("WARNING")
		case v >= slog.LevelError:
			a.Value = slog.StringValue("ERROR")
		default:
			a.Value = slog.StringValue("ERROR")
		}

	case slog.SourceKey:
		// Cleanup full system paths to be shorter yet resolvable via IDE.
		value := a.Value.String()[2 : len(a.Value.String())-1]
		partials := strings.Split(value, " ")
		full := strings.Join(partials[1:], ":")

		cwd, err := os.Getwd()
		if err != nil {
			a.Value = slog.StringValue(fmt.Sprintf("file://%s", full))
			break
		}

		rel, e := filepath.Rel(cwd, full)
		if e != nil {
			a.Value = slog.StringValue(fmt.Sprintf("file://%s", full))
			break
		}

		a.Value = slog.StringValue(fmt.Sprintf("%s", rel))
	}

	return a
}
