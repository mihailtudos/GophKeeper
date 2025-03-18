package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger(outputPath, level, format string) (*slog.Logger, error) {
	var loggerHandler slog.Handler
	var out io.Writer

	if outputPath == "stdout" || outputPath == "" {
		out = os.Stdout
	} else {
		file, err := getLoggerFile(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger file: %w", err)
		}

		out = file
	}

	switch format {
	case "json":
		loggerHandler = slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level: getLevel(level),
		})
	default:
		loggerHandler = slog.NewTextHandler(out, &slog.HandlerOptions{
			Level: getLevel(level),
		})
	}

	return slog.New(loggerHandler), nil
}

func getLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getLoggerFile(outputPath string) (*os.File, error) {
	dir := filepath.Dir(outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create or open log file: %w", err)
	}

	return file, nil
}

func ErrAttr(err error) slog.Attr {
	return slog.String("error", err.Error())
}
