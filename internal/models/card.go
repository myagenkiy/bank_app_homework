package models

import (
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AccountID       uuid.UUID `json:"account_id" db:"account_id"`
	EncryptedNumber string    `json:"-" db:"encrypted_number"`
	ExpiryHash      string    `json:"-" db:"expiry_hash"`
	CVVHash         string    `json:"-" db:"cvv_hash"`
	HMAC            string    `json:"-" db:"hmac"`
	Last4           string    `json:"last4" db:"last4"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type CardResponse struct {
	ID        uuid.UUID `json:"id"`
	Last4     string    `json:"last4"`
	Expiry    string    `json:"expiry"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCardRequest struct {
	AccountID uuid.UUID `json:"account_id" validate:"required"`
}

type CardPaymentRequest struct {
	CardNumber string  `json:"card_number" validate:"required"`
	CVV        string  `json:"cvv" validate:"required,len=3"`
	Expiry     string  `json:"expiry" validate:"required,len=4"`
	Amount     float64 `json:"amount" validate:"required,gt=0"`
}
