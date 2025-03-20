package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func (h *Handler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
	h.Logger.DebugContext(r.Context(), "retrieving secrets", slog.String("user_id", userID))

	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeJSON(w, r, http.StatusBadRequest, "failed to read request body")
		return
	}

	var password MasterPassword
	if err = json.Unmarshal(data, &password); err != nil {
		h.writeJSON(w, r, http.StatusBadRequest, "failed to decode password")
		return
	}

	secrets, err := h.Services.SecretsService.GetUserSecrets(r.Context(), userID, password.MasterPassword)
	if err != nil {
		h.logError(r, err)
		h.writeJSON(w, r, http.StatusInternalServerError, "failed to get user secrets")
		return
	}

	responseSecrets := make([]SecretResponse, 0)
	for _, secret := range *secrets {
		encodedSecret, _ := h.jsonEncodeSecretData(&secret)
		responseSecrets = append(responseSecrets, *encodedSecret)
	}

	h.writeJSON(w, r, http.StatusOK, responseSecrets)
}
