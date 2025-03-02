package infrastructure

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/mihailtudos/gophkeeper/server/internal/config"
)

var (
	instance              *slog.Logger
	once                  sync.Once
	defaultLoggerPath     = "./store/logs"
	defaultLoggerFileName = "gophkeeper.log"
)

func MustNewLogger(cfg config.LoggerConfig) *slog.Logger {
	logger, err := NewLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return logger
}

func NewLogger(cfg config.LoggerConfig) (*slog.Logger, error) {
	var errLogger error

	if instance != nil {
		return instance, nil
	}

	once.Do(func() {
		var loggerHandler slog.Handler
		var output io.Writer

		if cfg.Output == "stdout" {
			output = os.Stdout
		} else {
			file, err := getLoggerFile(cfg.Output)
			if err != nil {
				errLogger = err
				return
			}

			output = file
		}

		if cfg.Format == "json" {
			loggerHandler = slog.NewJSONHandler(output, &slog.HandlerOptions{
				Level: getLevel(cfg.Level),
			})
		}

		if cfg.Format == "text" {
			loggerHandler = slog.NewTextHandler(output, &slog.HandlerOptions{
				Level: getLevel(cfg.Level),
			})
		}

		instance = slog.New(loggerHandler)
	})

	return instance, errLogger
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

func getLoggerFile(p string) (*os.File, error) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create log directory %w", err)
		}

		return os.OpenFile(path.Join(p, defaultLoggerFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	return os.OpenFile(path.Join(defaultLoggerPath, defaultLoggerFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
