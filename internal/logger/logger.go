package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Log is the global structured logger instance
var Log *slog.Logger

func init() {
	// Initialize default logger to stderr with info level
	Log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Config defines logging configuration
type Config struct {
	Level    string // debug, info, warn, error
	Format   string // json, text, pretty
	FilePath string // optional file path for writing logs
}

// Setup initializes the global logger based on config
func Setup(cfg Config) (*os.File, error) {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var writer io.Writer = os.Stdout
	var logFile *os.File

	if cfg.FilePath != "" {
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			logFile = f
			writer = io.MultiWriter(os.Stdout, f)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: failed to open log file %s: %v\n", cfg.FilePath, err)
		}
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
		})
	} else if strings.ToLower(cfg.Format) == "pretty" {
		handler = &prettyHandler{
			writer: writer,
			level:  level,
		}
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)

	return logFile, nil
}

// prettyHandler provides clean, human-readable terminal output
type prettyHandler struct {
	writer io.Writer
	level  slog.Level
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	levelStr := r.Level.String()
	var color string
	switch r.Level {
	case slog.LevelDebug:
		color = "\033[36m" // Cyan
	case slog.LevelInfo:
		color = "\033[32m" // Green
	case slog.LevelWarn:
		color = "\033[33m" // Yellow
	case slog.LevelError:
		color = "\033[31m" // Red
	}
	reset := "\033[0m"

	timestamp := r.Time.Format(time.TimeOnly)
	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		return true
	})

	_, err := fmt.Fprintf(h.writer, "%s %s%-5s%s %s%s\n",
		timestamp, color, levelStr, reset, r.Message, attrs)
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return h
}
