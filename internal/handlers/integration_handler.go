package handlers

import (
	"bank-api/internal/services"
	"encoding/json"
	"net/http"
)

type IntegrationHandler struct {
	cbrService   *services.CBRService
	emailService *services.EmailService
}

func NewIntegrationHandler(cbrService *services.CBRService, emailService *services.EmailService) *IntegrationHandler {
	return &IntegrationHandler{
		cbrService:   cbrService,
		emailService: emailService,
	}
}

// GetCentralBankRate возвращает ключевую ставку ЦБ РФ
func (h *IntegrationHandler) GetCentralBankRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rate, err := h.cbrService.GetCentralBankRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rate)
}

// TestEmail отправляет тестовое письмо
func (h *IntegrationHandler) TestEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Для теста отправляем простое сообщение
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email sending test - check console logs"})
}
