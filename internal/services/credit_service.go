package services

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

type CreditService struct {
	creditRepo      *repositories.CreditRepository
	scheduleRepo    *repositories.PaymentScheduleRepository
	accountRepo     *repositories.AccountRepository
	transferService *TransferService
}

func NewCreditService(
	creditRepo *repositories.CreditRepository,
	scheduleRepo *repositories.PaymentScheduleRepository,
	accountRepo *repositories.AccountRepository,
	transferService *TransferService,
) *CreditService {
	return &CreditService{
		creditRepo:      creditRepo,
		scheduleRepo:    scheduleRepo,
		accountRepo:     accountRepo,
		transferService: transferService,
	}
}

// CalculateAnnuityPayment рассчитывает аннуитетный платеж
func (s *CreditService) CalculateAnnuityPayment(amount float64, annualRate float64, months int) float64 {
	monthlyRate := annualRate / 12 / 100
	if monthlyRate == 0 {
		return amount / float64(months)
	}
	annuity := monthlyRate * math.Pow(1+monthlyRate, float64(months)) / (math.Pow(1+monthlyRate, float64(months)) - 1)
	return amount * annuity
}

// GenerateSchedule генерирует график платежей
func (s *CreditService) GenerateSchedule(creditID uuid.UUID, amount float64, annualRate float64, months int) ([]*models.PaymentSchedule, error) {
	monthlyPayment := s.CalculateAnnuityPayment(amount, annualRate, months)
	var schedules []*models.PaymentSchedule

	dueDate := time.Now().AddDate(0, 1, 0) // первый платеж через месяц

	for i := 0; i < months; i++ {
		schedule := &models.PaymentSchedule{
			CreditID: creditID,
			DueDate:  dueDate.AddDate(0, i, 0),
			Amount:   monthlyPayment,
			Status:   models.PaymentPending,
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// ApplyForCredit оформление кредита
func (s *CreditService) ApplyForCredit(ctx context.Context, userID uuid.UUID, req *models.ApplyCreditRequest) (*models.CreditResponse, error) {
	// Создаем кредит
	credit := &models.Credit{
		UserID:       userID,
		Amount:       req.Amount,
		InterestRate: req.InterestRate,
		Status:       models.CreditActive,
	}

	err := s.creditRepo.Create(ctx, credit)
	if err != nil {
		return nil, err
	}

	// Генерируем график платежей
	schedules, err := s.GenerateSchedule(credit.ID, req.Amount, req.InterestRate, req.Months)
	if err != nil {
		return nil, err
	}

	err = s.scheduleRepo.CreateBatch(ctx, schedules)
	if err != nil {
		return nil, err
	}

	// Зачисляем сумму кредита на счет пользователя
	// Находим RUB счет пользователя
	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var rubAccount *models.Account
	for _, acc := range accounts {
		if acc.Currency == "RUB" {
			rubAccount = acc
			break
		}
	}

	if rubAccount == nil {
		return nil, errors.New("RUB account not found")
	}

	// Зачисляем деньги
	depositReq := &models.DepositRequest{
		AccountID:   rubAccount.ID,
		Amount:      req.Amount,
		Description: "Кредитные средства",
	}
	_, err = s.transferService.Deposit(ctx, depositReq)
	if err != nil {
		return nil, err
	}

	monthlyPayment := s.CalculateAnnuityPayment(req.Amount, req.InterestRate, req.Months)

	return &models.CreditResponse{
		ID:             credit.ID,
		Amount:         credit.Amount,
		InterestRate:   credit.InterestRate,
		MonthlyPayment: monthlyPayment,
		Status:         credit.Status,
		CreatedAt:      credit.CreatedAt,
	}, nil
}

// GetCreditSchedule получение графика платежей
func (s *CreditService) GetCreditSchedule(ctx context.Context, creditID uuid.UUID, userID uuid.UUID) ([]*models.PaymentScheduleResponse, error) {
	credit, err := s.creditRepo.FindByID(ctx, creditID)
	if err != nil || credit.UserID != userID {
		return nil, errors.New("credit not found or access denied")
	}

	schedules, err := s.scheduleRepo.FindByCreditID(ctx, creditID)
	if err != nil {
		return nil, err
	}

	var responses []*models.PaymentScheduleResponse
	for _, sched := range schedules {
		responses = append(responses, &models.PaymentScheduleResponse{
			ID:      sched.ID,
			DueDate: sched.DueDate.Format("2006-01-02"),
			Amount:  sched.Amount,
			Status:  sched.Status,
			PaidAt:  sched.PaidAt,
		})
	}
	return responses, nil
}
