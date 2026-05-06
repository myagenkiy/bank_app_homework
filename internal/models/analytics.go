package models

import (
	"time"
)

type MonthlyStats struct {
	Month            string  `json:"month"`
	TotalIncome      float64 `json:"total_income"`
	TotalExpense     float64 `json:"total_expense"`
	TransactionCount int     `json:"transaction_count"`
	NetChange        float64 `json:"net_change"`
}

type CreditBurden struct {
	TotalCreditAmount float64 `json:"total_credit_amount"`
	MonthlyPayments   float64 `json:"monthly_payments"`
	OverdueAmount     float64 `json:"overdue_amount"`
	NextPaymentDate   string  `json:"next_payment_date,omitempty"`
	RecommendedLimit  float64 `json:"recommended_limit"`
	CreditUtilization float64 `json:"credit_utilization"` // процент использования кредитного лимита
}

type BalancePrediction struct {
	CurrentBalance   float64   `json:"current_balance"`
	PredictedBalance float64   `json:"predicted_balance"`
	Confidence       float64   `json:"confidence"` // 0-1
	Days             int       `json:"days"`
	PredictionDate   time.Time `json:"prediction_date"`
	AvgDailyIncome   float64   `json:"avg_daily_income"`
	AvgDailyExpense  float64   `json:"avg_daily_expense"`
	UpcomingPayments float64   `json:"upcoming_payments"`
}

type DashboardResponse struct {
	TotalBalance   float64        `json:"total_balance"`
	TotalAccounts  int            `json:"total_accounts"`
	ActiveCredits  int            `json:"active_credits"`
	TotalCards     int            `json:"total_cards"`
	MonthlyStats   *MonthlyStats  `json:"monthly_stats"`
	CreditBurden   *CreditBurden  `json:"credit_burden"`
	RecentActivity []*Transaction `json:"recent_activity,omitempty"`
}
