package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Valentin-Makurin/gophermart/internal/auth"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	"github.com/Valentin-Makurin/gophermart/internal/service"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtManager  *auth.JWTManager
}

func NewAuthHandler(authService *service.AuthService, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		switch err {
		case models.ErrUserAlreadyExists:
			http.Error(w, "Login already taken", http.StatusConflict)
		default:
			log.Printf("Registration error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		switch err {
		case models.ErrUserNotFound, models.ErrInvalidPassword:
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
