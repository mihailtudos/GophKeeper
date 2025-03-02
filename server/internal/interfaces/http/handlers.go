package http

import (
	"encoding/json"
	"log/slog"
	nethttp "net/http"

	"github.com/mihailtudos/gophkeeper/server/internal/application/services"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
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

func (h *Handler) Register(w nethttp.ResponseWriter, r *nethttp.Request) {
	h.Logger.Info("register handler called")
	// TODO: implement

	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.Logger.Error("failed to decode request body", "error", err)
		w.WriteHeader(nethttp.StatusBadRequest)
		return
	}

	token, err := h.Services.AuthService.Register(r.Context(), user)
	if err != nil {
		h.Logger.Error("failed to register user", "error", err)
		w.WriteHeader(nethttp.StatusInternalServerError)
		return
	}

	w.Write([]byte("register handler called " + token))
	w.WriteHeader(nethttp.StatusOK)
}
