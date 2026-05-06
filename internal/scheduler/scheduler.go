package scheduler

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"bank-api/internal/services"
	"context"
	"log"
	"time"
)

type Scheduler struct {
	paymentScheduleRepo *repositories.PaymentScheduleRepository
	creditRepo          *repositories.CreditRepository
	accountRepo         *repositories.AccountRepository
	transferService     *services.TransferService
}

func NewScheduler(
	paymentScheduleRepo *repositories.PaymentScheduleRepository,
	creditRepo *repositories.CreditRepository,
	accountRepo *repositories.AccountRepository,
	transferService *services.TransferService,
) *Scheduler {
	return &Scheduler{
		paymentScheduleRepo: paymentScheduleRepo,
		creditRepo:          creditRepo,
		accountRepo:         accountRepo,
		transferService:     transferService,
	}
}

// Start запускает шедулер (каждые 12 часов)
func (s *Scheduler) Start(ctx context.Context) {
	log.Println("🕐 Scheduler started - checking overdue payments every 12 hours")

	// Запускаем сразу при старте
	s.checkOverduePayments(ctx)

	// Затем каждые 12 часов
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkOverduePayments(ctx)
		case <-ctx.Done():
			log.Println("🛑 Scheduler stopped")
			return
		}
	}
}

// checkOverduePayments проверяет и обрабатывает просроченные платежи
func (s *Scheduler) checkOverduePayments(ctx context.Context) {
	log.Println("🔍 Checking overdue payments...")

	overduePayments, err := s.paymentScheduleRepo.FindOverduePayments(ctx, time.Now())
	if err != nil {
		log.Printf("❌ Failed to find overdue payments: %v", err)
		return
	}

	if len(overduePayments) == 0 {
		log.Println("✅ No overdue payments found")
		return
	}

	log.Printf("⚠️ Found %d overdue payments", len(overduePayments))

	for _, payment := range overduePayments {
		s.processOverduePayment(ctx, payment)
	}
}

// processOverduePayment обрабатывает один просроченный платеж
func (s *Scheduler) processOverduePayment(ctx context.Context, payment *models.PaymentSchedule) {
	log.Printf("💰 Processing overdue payment: %s, amount: %.2f", payment.ID, payment.Amount)

	credit, err := s.creditRepo.FindByID(ctx, payment.CreditID)
	if err != nil {
		log.Printf("❌ Failed to find credit: %v", err)
		return
	}

	accounts, err := s.accountRepo.FindByUserID(ctx, credit.UserID)
	if err != nil {
		log.Printf("❌ Failed to find user accounts: %v", err)
		return
	}

	var rubAccount *models.Account
	for _, acc := range accounts {
		if acc.Currency == "RUB" {
			rubAccount = acc
			break
		}
	}

	if rubAccount == nil {
		log.Printf("❌ No RUB account found for user %s", credit.UserID)
		return
	}

	err = s.transferService.ProcessScheduledPayment(ctx, rubAccount.ID, payment.Amount,
		"Ежемесячный платеж по кредиту")

	if err == nil {
		now := time.Now()
		err = s.paymentScheduleRepo.UpdateStatus(ctx, payment.ID, models.PaymentPaid, &now)
		if err != nil {
			log.Printf("❌ Failed to update payment status: %v", err)
			return
		}
		log.Printf("✅ Payment %s processed successfully", payment.ID)
	} else {
		log.Printf("⚠️ Insufficient funds for payment %s, applying penalty", payment.ID)

		penaltyAmount := payment.Amount * 0.1
		payment.Amount += penaltyAmount

		err = s.paymentScheduleRepo.UpdateStatus(ctx, payment.ID, models.PaymentOverdue, nil)
		if err != nil {
			log.Printf("❌ Failed to update payment status to overdue: %v", err)
			return
		}

		err = s.creditRepo.UpdateStatus(ctx, credit.ID, models.CreditOverdue)
		if err != nil {
			log.Printf("❌ Failed to update credit status: %v", err)
			return
		}

		log.Printf("⚠️ Penalty applied: +%.2f, new amount: %.2f", penaltyAmount, payment.Amount)
	}
}
