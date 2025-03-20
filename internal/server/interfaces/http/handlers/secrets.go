package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/internal/server/application/services/secrets"
	"github.com/mihailtudos/gophkeeper/internal/server/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type MasterPassword struct {
	MasterPassword string `json:"master_password"`
}

type SecretResponse struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Store handles the POST request to store a secret.
func (h *Handler) Store(w http.ResponseWriter, r *http.Request) {
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
		h.errorResponse(w, r, http.StatusBadRequest, "failed to read request body")
		return
	}

	defer r.Body.Close()

	requestData := SecretRequest{}
	if err = json.Unmarshal(body, &requestData); err != nil {
		h.errorResponse(w, r, http.StatusBadRequest, "failed to unmarshal request body")
		return
	}

	if err = h.Services.SecretsService.StoreSecret(r.Context(), userID, requestData.Type,
		requestData.Name, requestData.MasterPassword, requestData.Data); err != nil {
		if errors.Is(err, secrets.ErrInvalidSecretType) {
			h.errorResponse(w, r, http.StatusBadRequest, "invalid secret type")
			return
		}

		if errors.Is(err, secrets.ErrSecretExists) {
			h.errorResponse(w, r, http.StatusBadRequest, "secret already exists")
			return
		}

		if errors.Is(err, repositories.ErrRecordNotFound) {
			h.errorResponse(w, r, http.StatusBadRequest, "user not found")
			return
		}

		h.logError(r, err)
		h.errorResponse(w, r, http.StatusInternalServerError, "failed to store secret")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DecryptSecret handles the GET request to decrypt a secret.
func (h *Handler) DecryptSecret(w http.ResponseWriter, r *http.Request) {
	secretID := chi.URLParam(r, "id")
	if secretID == "" {
		h.errorResponse(w, r, http.StatusBadRequest, "missing secret id")
		return
	}

	var password MasterPassword
	if err := json.NewDecoder(r.Body).Decode(&password); err != nil || password.MasterPassword == "" {
		h.errorResponse(w, r, http.StatusBadRequest, "invalid master password")
		return
	}

	h.Logger.DebugContext(r.Context(), "fetching secret by id", slog.String("secret_id", secretID))

	secret, err := h.Services.SecretsService.GetSecretByID(r.Context(), secretID, password.MasterPassword)
	if err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			h.errorResponse(w, r, http.StatusNotFound, "secret not found")
			return
		}
	}

	encodedSecret, _ := h.jsonEncodeSecretData(secret)

	h.writeJSON(w, r, http.StatusOK, encodedSecret)
}

func (h *Handler) jsonEncodeSecretData(secret *domain.Secret) (*SecretResponse, error) {
	var data interface{}

	switch secret.SType {
	case "login":
		var l domain.LoginSecret
		if err := json.Unmarshal(secret.Data, &l); err != nil {
			h.Logger.Error("failed to unmarshal login secret", logger.ErrAttr(err))
			return nil, fmt.Errorf("failed to unmarshal login secret: %w", err)
		}
		data = l

	case "card":
		var c domain.CardDetails
		if err := json.Unmarshal(secret.Data, &c); err != nil {
			h.Logger.Error("failed to unmarshal card details", logger.ErrAttr(err))
			return nil, fmt.Errorf("failed to unmarshal card details: %w", err)
		}

		data = c
	case "text":
		var t domain.PlainText
		if err := json.Unmarshal(secret.Data, &t); err != nil {
			h.Logger.Error("failed to unmarshal plain text", logger.ErrAttr(err))
			return nil, fmt.Errorf("failed to unmarshal plain text: %w", err)
		}

		data = t
	case "binary":
		binaryData := domain.BinaryData{Data: secret.Data}
		data = binaryData.ToBase64() // Encode binary data as Base64

	default:
		h.Logger.Error("unknown secret type", slog.String("secret_type", secret.SType))
		return nil, errors.New("unknown secret type")
	}

	return &SecretResponse{
		ID:        secret.ID,
		Type:      secret.SType,
		Name:      secret.SName,
		Data:      data,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}, nil
}
