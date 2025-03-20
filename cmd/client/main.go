package main

import (
	"context"
	"github.com/mihailtudos/gophkeeper/internal/client"
	"github.com/mihailtudos/gophkeeper/internal/client/application/security"
	"github.com/mihailtudos/gophkeeper/internal/client/application/services"
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
	lg, err := logger.NewLogger(cfg.Logger.OutputPath, cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		panic(err)
	}

	lg.Info("Starting GophKeeper client", "version", buildVersion, "build_date", buildDate)
	lg.Debug("Debug mode enabled")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	km := security.NewKeyManager()

	s := services.NewServices(ctx, lg, cfg, km)

	app := client.NewApp(ctx, cfg, lg, s)
	//m := NewModel(cfg)
	stop := make(chan os.Signal, 1)
	go app.Run(ctx, stop)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sign := <-stop
	lg.Debug("stopping application", slog.String("signal", sign.String()))

	app.Stop()

	lg.Debug("application stopped")
}
