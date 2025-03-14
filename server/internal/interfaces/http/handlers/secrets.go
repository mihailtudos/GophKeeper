package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mihailtudos/gophkeeper/server/internal/application/services/secrets"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type SecretResponse struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (h *Handler) Store(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	userID := r.Header.Get("user_id")

	// Generic request struct
	type SecretRequest struct {
		Type           string          `json:"type"`
		Name           string          `json:"name"`
		Data           json.RawMessage `json:"data"`
		MasterPassword string          `json:"master_password"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error("failed to read request body", pkg.ErrAttr(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.Body.Close()

	requestData := SecretRequest{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		h.Logger.Error("failed to unmarshal request body", pkg.ErrAttr(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.Services.SecretsService.StoreSecret(r.Context(), userID, requestData.Type, requestData.Name, requestData.MasterPassword, requestData.Data); err != nil {
		if errors.Is(err, secrets.ErrInvalidSecretType) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid secret type"))
			return
		}

		if errors.Is(err, repositories.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("user not found"))
			return
		}

		h.Logger.Error("failed to store secret", pkg.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write([]byte(requestData.Type)); err != nil {
		h.Logger.Error("failed to write response", pkg.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DecryptSecret(w http.ResponseWriter, r *http.Request) {
	secretID := chi.URLParam(r, "id")
	if secretID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type masterPassword struct {
		MasterPassword string `json:"master_password"`
	}

	var password masterPassword
	if err := json.NewDecoder(r.Body).Decode(&password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if password.MasterPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Logger.DebugContext(r.Context(), "fetching secret by id", slog.String("secret_id", secretID))

	secret, err := h.Services.SecretsService.GetSecretByID(r.Context(), secretID, password.MasterPassword)
	if err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encodedSecret, _ := h.jsonEncodeSecret(secret)

	if _, err = w.Write(encodedSecret); err != nil {
		h.Logger.Error("failed to encode secret", pkg.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) jsonEncodeSecret(secret *domain.Secret) ([]byte, error) {
	var data interface{}

	switch secret.SType {
	case "login":
		var l domain.LoginSecret
		if err := json.Unmarshal(secret.Data, &l); err != nil {
			h.Logger.Error("failed to unmarshal login secret", pkg.ErrAttr(err))
			return nil, fmt.Errorf("failed to unmarshal login secret: %w", err)
		}
		data = l

	case "card":
		var c domain.CardDetails
		if err := json.Unmarshal(secret.Data, &c); err != nil {
			h.Logger.Error("failed to unmarshal card details", pkg.ErrAttr(err))
			return nil, fmt.Errorf("failed to unmarshal card details: %w", err)
		}
		data = c

	case "text":
		data = string(secret.Data) // Plain text stored directly as a string

	case "binary":
		binaryData := domain.BinaryData{Data: secret.Data}
		data = binaryData.StringBase64() // Encode binary data as Base64

	default:
		h.Logger.Error("unknown secret type", slog.String("secret_type", secret.SType))
		return nil, errors.New("unknown secret type")
	}

	// Unified Response
	response := SecretResponse{
		ID:        secret.ID,
		Type:      secret.SType,
		Name:      secret.SName,
		Data:      data,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}

	return json.Marshal(response)
}
