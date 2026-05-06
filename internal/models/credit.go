package models

import (
	"time"

	"github.com/google/uuid"
)

type CreditStatus string

const (
	CreditActive  CreditStatus = "active"
	CreditClosed  CreditStatus = "closed"
	CreditOverdue CreditStatus = "overdue"
)

type Credit struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	UserID       uuid.UUID    `json:"user_id" db:"user_id"`
	Amount       float64      `json:"amount" db:"amount"`
	InterestRate float64      `json:"interest_rate" db:"interest_rate"` // годовая, %
	Status       CreditStatus `json:"status" db:"status"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
}

type ApplyCreditRequest struct {
	Amount       float64 `json:"amount" validate:"required,gt=0,lte=1000000"`
	InterestRate float64 `json:"interest_rate" validate:"required,min=1,max=30"`
	Months       int     `json:"months" validate:"required,min=3,max=60"` // срок кредита в месяцах
}

type CreditResponse struct {
	ID             uuid.UUID    `json:"id"`
	Amount         float64      `json:"amount"`
	InterestRate   float64      `json:"interest_rate"`
	MonthlyPayment float64      `json:"monthly_payment"`
	Status         CreditStatus `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
}
