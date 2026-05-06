package services

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"bank-api/internal/utils"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type CardService struct {
	cardRepo    *repositories.CardRepository
	accountRepo *repositories.AccountRepository
}

func NewCardService(cardRepo *repositories.CardRepository, accountRepo *repositories.AccountRepository) *CardService {
	return &CardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
	}
}

func (s *CardService) CreateCard(ctx context.Context, accountID, userID uuid.UUID) (*models.CardResponse, error) {
	// Проверяем, что счет принадлежит пользователю
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil || account.UserID != userID {
		return nil, errors.New("account not found or access denied")
	}

	// Генерируем данные карты
	cardNumber := utils.GenerateCardNumber()
	expiry := utils.GenerateExpiry()

	// Генерируем CVV (3 случайные цифры)
	cvv := fmt.Sprintf("%03d", int(utils.GenerateCardNumber()[0])%1000)

	// Шифруем номер карты
	encryptedNumber, err := utils.EncryptCardData(cardNumber + "|" + expiry)
	if err != nil {
		return nil, err
	}

	// Хешируем expiry и CVV
	expiryHash := utils.HashExpiry(expiry)
	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		return nil, err
	}

	// Вычисляем HMAC для целостности
	hmacValue := utils.ComputeHMAC(encryptedNumber + expiryHash + cvvHash)

	card := &models.Card{
		AccountID:       accountID,
		EncryptedNumber: encryptedNumber,
		ExpiryHash:      expiryHash,
		CVVHash:         cvvHash,
		HMAC:            hmacValue,
		Last4:           cardNumber[len(cardNumber)-4:],
	}

	err = s.cardRepo.Create(ctx, card)
	if err != nil {
		return nil, err
	}

	return &models.CardResponse{
		ID:        card.ID,
		Last4:     card.Last4,
		Expiry:    expiry,
		CreatedAt: card.CreatedAt,
	}, nil
}

func (s *CardService) GetUserCards(ctx context.Context, accountID, userID uuid.UUID) ([]*models.CardResponse, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil || account.UserID != userID {
		return nil, errors.New("account not found or access denied")
	}

	cards, err := s.cardRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var responses []*models.CardResponse
	for _, card := range cards {
		responses = append(responses, &models.CardResponse{
			ID:        card.ID,
			Last4:     card.Last4,
			Expiry:    "****",
			CreatedAt: card.CreatedAt,
		})
	}
	return responses, nil
}

func (s *CardService) MakePayment(ctx context.Context, req *models.CardPaymentRequest) error {
	// Расшифровываем и проверяем карту (упрощенно)
	if !utils.ValidateCardNumber(req.CardNumber) {
		return errors.New("invalid card number")
	}
	return nil
}
