package main

import (
	"context"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/server/application/services"
	"github.com/mihailtudos/gophkeeper/internal/server/config"
	"github.com/mihailtudos/gophkeeper/internal/server/infrastructure/db"
	"github.com/mihailtudos/gophkeeper/internal/server/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/internal/server/interfaces/http/handlers"
	mw "github.com/mihailtudos/gophkeeper/internal/server/interfaces/http/middleware"
	l "github.com/mihailtudos/gophkeeper/pkg/logger"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	cfg := config.MustNewConfig()

	Logger, err := l.NewLogger(cfg.Logger.OutputPath, cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		panic(err)
	}

	Logger.Info("Logger initialized successfully")
	Logger.Debug("debug mode enabled")

	fmt.Printf("config: %+v\n", cfg)

	ctx := context.Background()

	DB, err := db.NewDB(ctx, Logger, cfg.Database.DSN)
	if err != nil {
		Logger.Error("failed to connect to database", "error", err)
		return
	}

	defer DB.Close()
	Logger.Info("database connected successfully")

	repository, err := repositories.NewRepository(ctx, DB, Logger)
	if err != nil {
		Logger.Error("failed to create repository", "error", err)
		log.Fatal("failed to create repository", "error", err)
	}

	ss := services.NewServices(ctx, cfg, Logger, repository)

	h := handlers.NewHandler(Logger, ss)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	router.Route("/api", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.Auth.SecretKey, Logger))
			r.Route("/secrets", func(r chi.Router) {
				r.Post("/", h.Store)
				r.Post("/{id}/decrypt", h.DecryptSecret)
			})
			r.Route("/users", func(r chi.Router) {
				r.Post("/secrets", h.GetSecrets)
			})
		})
	})

	srv := http.Server{
		Addr: fmt.Sprintf("%s:%s",
			cfg.HTTPServer.Host,
			cfg.HTTPServer.Port),
		Handler: router,
	}

	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		Logger.Error("failed to start server", "error", err)
		log.Fatal("failed to start server", "error", err)
	}
}
