package models

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransferTransaction TransactionType = "transfer"
	DepositTransaction  TransactionType = "deposit"
	WithdrawTransaction TransactionType = "withdraw"
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id" db:"id"`
	FromAccountID *uuid.UUID        `json:"from_account_id,omitempty" db:"from_account_id"`
	ToAccountID   *uuid.UUID        `json:"to_account_id,omitempty" db:"to_account_id"`
	Amount        float64           `json:"amount" db:"amount"`
	Type          TransactionType   `json:"type" db:"type"`
	Status        TransactionStatus `json:"status" db:"status"`
	Description   string            `json:"description" db:"description"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
}

type TransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	Description   string    `json:"description"`
}

type DepositRequest struct {
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description"`
}

type WithdrawRequest struct {
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description"`
}
