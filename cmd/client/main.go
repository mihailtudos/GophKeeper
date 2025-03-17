package main

import (
	"context"
	"github.com/mihailtudos/gophkeeper/internal/client"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	cfg := config.MustNewConfig()
	Logger, err := logger.NewLogger(cfg.Logger.OutputPath, cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		panic(err)
	}

	Logger.Info("Starting GophKeeper client", "version", buildVersion, "build_date", buildDate)
	Logger.Debug("Debug mode enabled")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := client.NewApp(ctx, cfg, Logger)
	//m := NewModel(cfg)
	stop := make(chan os.Signal, 1)
	go app.Run(ctx, stop)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sign := <-stop
	Logger.Debug("stopping application", slog.String("signal", sign.String()))

	app.Stop()

	Logger.Debug("application stopped")
}
