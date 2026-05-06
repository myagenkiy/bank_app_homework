package models

type CentralBankRate struct {
	Rate       float64 `json:"rate"`
	BankMargin float64 `json:"bank_margin"`
	FinalRate  float64 `json:"final_rate"`
	UpdatedAt  string  `json:"updated_at"`
}

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type PaymentNotification struct {
	UserEmail   string  `json:"user_email"`
	UserName    string  `json:"user_name"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // credit_payment, transfer, deposit
	Status      string  `json:"status"`
	Description string  `json:"description"`
}
