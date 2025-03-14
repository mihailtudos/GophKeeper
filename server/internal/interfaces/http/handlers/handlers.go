package handlers

import (
	"github.com/mihailtudos/gophkeeper/server/internal/application/services"
	"log/slog"
)

type Handler struct {
	Logger   *slog.Logger
	Services *services.Services
}

func NewHandler(logger *slog.Logger, services *services.Services) *Handler {
	return &Handler{
		Logger:   logger,
		Services: services,
	}
}
