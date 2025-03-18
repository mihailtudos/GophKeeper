package handlers

import (
	"encoding/json"
	"errors"
	"github.com/mihailtudos/gophkeeper/internal/server/application/services/auth"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"net/http"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var registrationRequest auth.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&registrationRequest); err != nil {
		h.Logger.Error("failed to decode request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokens, err := h.Services.AuthService.Register(r.Context(), registrationRequest)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("user already exists"))
			return
		}

		h.Logger.Error("failed to register user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		h.Logger.Error("failed to marshal response", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest auth.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		h.Logger.Error("failed to decode request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Logger.Debug("login request", "username", loginRequest.Username)

	tokens, err := h.Services.AuthService.Login(r.Context(), loginRequest)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("user not found"))
			return
		}

		if errors.Is(err, auth.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("user not found"))
			return
		}

		h.Logger.Error("failed to login user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		h.Logger.Error("failed to marshal response", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Extract refresh token from request
	var requestData struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.RefreshToken == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := h.Services.AuthService.RefreshToken(r.Context(), requestData.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) || errors.Is(err, auth.ErrTokenNotFound) {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		h.Logger.Error("failed to refresh token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		h.Logger.Error("failed to marshal response", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(data); err != nil {
		h.Logger.Error("failed to write response", logger.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
