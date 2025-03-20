package handlers

import (
	"encoding/json"
	"errors"
	"github.com/mihailtudos/gophkeeper/internal/server/application/services/auth"
	"net/http"
)

// Register handles user registration.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var registrationRequest auth.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&registrationRequest); err != nil {
		h.errorResponse(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.Services.AuthService.Register(r.Context(), registrationRequest)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			h.errorResponse(w, r, http.StatusConflict, "user already exists")
			return
		}

		h.logError(r, err)
		h.errorResponse(w, r, http.StatusInternalServerError, "failed to register user")
		return
	}

	h.writeJSON(w, r, http.StatusOK, tokens)
}

// Login handles user login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest auth.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		h.errorResponse(w, r, http.StatusBadRequest, "failed to decode request body")
		return
	}

	h.Logger.Debug("login request", "username", loginRequest.Username)

	tokens, err := h.Services.AuthService.Login(r.Context(), loginRequest)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			h.errorResponse(w, r, http.StatusNotFound, "user not found")
			return
		}

		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.errorResponse(w, r, http.StatusUnauthorized, "not authorized")
			return
		}

		h.logError(r, err)
		h.errorResponse(w, r, http.StatusInternalServerError, "failed to login user")
		return
	}

	h.writeJSON(w, r, http.StatusOK, tokens)
}

// Refresh handles the refresh token request.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Extract refresh token from request
	var requestData struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.RefreshToken == "" {
		h.errorResponse(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	tokens, err := h.Services.AuthService.RefreshToken(r.Context(), requestData.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) || errors.Is(err, auth.ErrTokenNotFound) {
			h.errorResponse(w, r, http.StatusUnauthorized, "invalid refresh token")
			return
		}

		h.logError(r, err)
		h.errorResponse(w, r, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	h.writeJSON(w, r, http.StatusOK, tokens)
}
