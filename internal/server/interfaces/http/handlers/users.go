package handlers

import (
	"encoding/json"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"io"
	"log/slog"
	"net/http"
)

func (h *Handler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	h.Logger.DebugContext(r.Context(), "fetching secrets")
	userID := r.Header.Get("user_id")
	h.Logger.DebugContext(r.Context(), "fetching secrets", slog.String("user_id", userID))

	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error("failed to read request body", logger.ErrAttr(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var password MasterPassword
	if err := json.Unmarshal(data, &password); err != nil {
		h.Logger.Error("failed to decode password", logger.ErrAttr(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secrets, err := h.Services.SecretsService.GetUserSecrets(r.Context(), userID, password.MasterPassword)
	if err != nil {
		h.Logger.Error("failed to get user secrets", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseSecrets := make([]SecretResponse, 0)
	for _, secret := range *secrets {
		encodedSecret, _ := h.jsonEncodeSecretData(&secret)
		responseSecrets = append(responseSecrets, *encodedSecret)
	}

	data, err = json.Marshal(responseSecrets)
	if err != nil {
		h.Logger.Error("failed to marshal secrets", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(data); err != nil {
		h.Logger.Error("failed to write response", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
