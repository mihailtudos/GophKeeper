package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/middleware"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"log/slog"
	"net/http"
)

var (
	requestIDKey = "request_id"
)

func (h *Handler) writeJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		h.logError(r, err)
		h.errorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)
}

func (h *Handler) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	envelope := map[string]interface{}{"error": message}

	data, err := json.Marshal(envelope)
	if err != nil {
		h.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(data); err != nil {
		h.logError(r, err)
	}
}

func (h *Handler) logError(r *http.Request, err error) {
	h.Logger.Error("server error",
		slog.String("request_id", middleware.GetReqID(r.Context())),
		slog.String("url", r.URL.String()),
		slog.String("method", r.Method),
		logger.ErrAttr(err))
}
