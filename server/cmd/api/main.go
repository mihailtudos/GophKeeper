package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mihailtudos/gophkeeper/server/internal/application/services"
	"github.com/mihailtudos/gophkeeper/server/internal/config"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/db"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
	h "github.com/mihailtudos/gophkeeper/server/internal/interfaces/http"
)

func main() {
	cfg := config.NewConfig()

	logger := infrastructure.MustNewLogger(cfg.Logger)

	logger.Info("logger initialized successfully")
	logger.Debug("debug mode enabled")

	ctx := context.Background()
	// TODO: add db
	db, err := db.NewDB(ctx, logger, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return
	}

	defer db.Close()
	logger.Info("database connected successfully")

	// TODO: add repositories
	repository, err := repositories.NewRepository(ctx, db, logger)
	if err != nil {
		logger.Error("failed to create repository", "error", err)
		log.Fatal("failed to create repository", "error", err)
	}

	// TODO: add server
	services := services.NewServices(ctx, cfg, logger, repository)

	// TODO: add handlers
	handlers := h.NewHandler(logger, services)

	// TODO: add router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	// TODO: add handlers
	router.Route("/api", func(r chi.Router) {
		router.Post("/signup", handlers.Register)
	})

	// TODO: add server
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err)
		log.Fatal("failed to start server", "error", err)
	}
}
