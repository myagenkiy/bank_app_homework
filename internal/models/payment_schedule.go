package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
	PaymentOverdue PaymentStatus = "overdue"
)

type PaymentSchedule struct {
	ID       uuid.UUID     `json:"id" db:"id"`
	CreditID uuid.UUID     `json:"credit_id" db:"credit_id"`
	DueDate  time.Time     `json:"due_date" db:"due_date"`
	Amount   float64       `json:"amount" db:"amount"`
	Status   PaymentStatus `json:"status" db:"status"`
	PaidAt   *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
}

type PaymentScheduleResponse struct {
	ID      uuid.UUID     `json:"id"`
	DueDate string        `json:"due_date"`
	Amount  float64       `json:"amount"`
	Status  PaymentStatus `json:"status"`
	PaidAt  *time.Time    `json:"paid_at,omitempty"`
}
