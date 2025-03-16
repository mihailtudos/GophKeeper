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
	"github.com/mihailtudos/gophkeeper/server/internal/interfaces/http/handlers"
	mw "github.com/mihailtudos/gophkeeper/server/internal/interfaces/http/middleware"
)

func main() {
	cfg := config.MustNewConfig()

	logger := infrastructure.MustNewLogger(cfg.Logger)

	logger.Info("logger initialized successfully")
	logger.Debug("debug mode enabled")

	fmt.Printf("config: %+v\n", cfg)

	ctx := context.Background()

	DB, err := db.NewDB(ctx, logger, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return
	}

	defer DB.Close()
	logger.Info("database connected successfully")

	repository, err := repositories.NewRepository(ctx, DB, logger)
	if err != nil {
		logger.Error("failed to create repository", "error", err)
		log.Fatal("failed to create repository", "error", err)
	}

	ss := services.NewServices(ctx, cfg, logger, repository)

	h := handlers.NewHandler(logger, ss)

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
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.Auth.SecretKey, logger))
			r.Route("/secrets", func(r chi.Router) {
				r.Post("/", h.Store)
				r.Post("/{id}/decrypt", h.DecryptSecret)
			})
			r.Route("/users", func(r chi.Router) {
				r.Post("/secrets", h.GetSecrets)
			})
		})
	})

	// TODO: add server
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,
	}

	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err)
		log.Fatal("failed to start server", "error", err)
	}
}
