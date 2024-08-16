package handlers

import (
	"encoding/json"
	"errors"
	"go-url-shortener/internal/storage"
	"go-url-shortener/internal/utils"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Handler struct {
	Storage *storage.Storage
}

type AuthHandler struct {
	Storage *storage.Storage
}

func (h *Handler) HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Url shoterer is ready to short you links! ;)"))
}

// CreateShortLinkHandler обрабатывает запросы на создание новой сокращенной ссылки.
func (h *Handler) CreateShortLinkHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Link string `json:"link"`
		Name string `json:"name,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.Link == "" {
		http.Error(w, "Link is required", http.StatusBadRequest)
		return
	}

	// Получаем email пользователя из токена
	email, err := h.getUserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if requestData.Name == "" {
		requestData.Name = utils.GenerateShortURL()
	}

	if h.Storage.NameExists(requestData.Name) {
		http.Error(w, "Link name already exists", http.StatusConflict)
		return
	}

	if err := h.Storage.SaveLink(requestData.Name, requestData.Link, email); err != nil {
		http.Error(w, "Failed to save link", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"short_link": requestData.Name,
	}
	json.NewEncoder(w).Encode(response)
}

// RedirectHandler перенаправляет пользователя на исходный URL по сокращенной ссылке.
func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortName := params["shortName"]

	link, err := h.Storage.GetLink(shortName)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, link, http.StatusMovedPermanently)
}

// GetUserFromRequest извлекает email пользователя из заголовка Authorization.
func (h *Handler) getUserFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	email, err := utils.VerifyToken(token)
	if err != nil {
		return "", errors.New("invalid token")
	}

	return email, nil
}

// RegisterHandler создает новый токен для пользователя.
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if h.Storage.EmailExists(requestData.Email) {
		http.Error(w, "Email already in use", http.StatusConflict)
		return
	}

	tokenString, err := utils.GenerateToken(requestData.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	if err := h.Storage.SaveToken(requestData.Email, tokenString); err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

// AdminDeleteUserHandler позволяет администратору удалять пользователя и его ссылки.
func (h *AuthHandler) AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var requestData struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.Storage.EmailExists(requestData.Email) {
		http.Error(w, "Email not exists", http.StatusConflict)
		return
	}

	if err := h.Storage.DeleteUser(requestData.Email); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminGetUsersHandler возвращает список всех пользователей и количество их ссылок.
func (h *AuthHandler) AdminGetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	users := h.Storage.GetAllUsers()
	json.NewEncoder(w).Encode(users)
}

// isAdmin проверяет, является ли запрос от администратора.
func (h *AuthHandler) isAdmin(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token == h.Storage.GetAdminToken()
}
