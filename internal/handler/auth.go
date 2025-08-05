package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/NOTMKW/JWT/internal/dto"
	"github.com/NOTMKW/JWT/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *service.Authservice
	validator   *validator.Validate
}

func NewAuthHandler(authService *service.Authservice) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed:"+err.Error())
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusConflict, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation Failed:"+err.Error())
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		h.writeErrorResponse(w, http.StatusUnauthorized, "bearer token required")
		return
	}
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid token")
		return
	}

	response := map[string]interface{}{
		"message": "Access Granted",
		"user": map[string]string{
			"id":    claims.UserID,
			"email": claims.Email,
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)

}

func (h *AuthHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, dto.ErrorResponse{Error: message})
}
