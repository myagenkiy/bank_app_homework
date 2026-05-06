package handlers

import (
	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"encoding/json"
	"net/http"
)

type AccountHandler struct {
	accountRepo *repositories.AccountRepository
}

// NewAccountHandler создает новый обработчик для счетов
func NewAccountHandler(accountRepo *repositories.AccountRepository) *AccountHandler {
	return &AccountHandler{accountRepo: accountRepo}
}

// CreateAccount создает новый банковский счет
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found in context"})
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Currency = "RUB"
	}

	if req.Currency == "" {
		req.Currency = "RUB"
	}

	account, err := h.accountRepo.Create(r.Context(), userID, req.Currency)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create account"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account created successfully",
		"account": account,
	})
}

// GetAccounts возвращает все счета пользователя
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found in context"})
		return
	}

	accounts, err := h.accountRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get accounts"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	})
}
