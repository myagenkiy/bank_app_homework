package services

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	accountRepo     *repositories.AccountRepository
	transactionRepo *repositories.TransactionRepository
	creditRepo      *repositories.CreditRepository
	cardRepo        *repositories.CardRepository
}

func NewAnalyticsService(
	accountRepo *repositories.AccountRepository,
	transactionRepo *repositories.TransactionRepository,
	creditRepo *repositories.CreditRepository,
	cardRepo *repositories.CardRepository,
) *AnalyticsService {
	return &AnalyticsService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		creditRepo:      creditRepo,
		cardRepo:        cardRepo,
	}
}

// GetDashboard получает сводную информацию для пользователя
func (s *AnalyticsService) GetDashboard(ctx context.Context, userID uuid.UUID) (*models.DashboardResponse, error) {
	// Получаем все счета
	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Считаем общий баланс
	var totalBalance float64
	for _, acc := range accounts {
		totalBalance += acc.Balance
	}

	// Получаем кредиты
	credits, err := s.creditRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	activeCredits := 0
	for _, c := range credits {
		if c.Status == models.CreditActive || c.Status == models.CreditOverdue {
			activeCredits++
		}
	}

	// Получаем карты
	var totalCards int
	for _, acc := range accounts {
		cards, err := s.cardRepo.FindByAccountID(ctx, acc.ID)
		if err != nil {
			continue
		}
		totalCards += len(cards)
	}

	// Статистика за текущий месяц
	now := time.Now()
	monthlyStats, err := s.transactionRepo.GetMonthlyStats(ctx, accounts[0].ID, now.Year(), int(now.Month()))
	if err != nil {
		monthlyStats = &models.MonthlyStats{
			Month:            now.Format("2006-01"),
			TotalIncome:      0,
			TotalExpense:     0,
			TransactionCount: 0,
			NetChange:        0,
		}
	} else {
		monthlyStats.NetChange = monthlyStats.TotalIncome - monthlyStats.TotalExpense
	}

	// Кредитная нагрузка
	creditBurden, err := s.GetCreditBurden(ctx, userID)
	if err != nil {
		creditBurden = &models.CreditBurden{
			TotalCreditAmount: 0,
			MonthlyPayments:   0,
			OverdueAmount:     0,
			RecommendedLimit:  0,
			CreditUtilization: 0,
		}
	}

	// Последние транзакции
	var recentActivity []*models.Transaction
	if len(accounts) > 0 {
		recentActivity, _ = s.transactionRepo.GetByAccountID(ctx, accounts[0].ID, 5)
	}

	return &models.DashboardResponse{
		TotalBalance:   totalBalance,
		TotalAccounts:  len(accounts),
		ActiveCredits:  activeCredits,
		TotalCards:     totalCards,
		MonthlyStats:   monthlyStats,
		CreditBurden:   creditBurden,
		RecentActivity: recentActivity,
	}, nil
}

// GetCreditBurden рассчитывает кредитную нагрузку
func (s *AnalyticsService) GetCreditBurden(ctx context.Context, userID uuid.UUID) (*models.CreditBurden, error) {
	credits, err := s.creditRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var totalAmount, monthlyPayments, overdueAmount float64
	for _, credit := range credits {
		if credit.Status == models.CreditActive || credit.Status == models.CreditOverdue {
			totalAmount += credit.Amount
			// Рассчитываем примерный ежемесячный платеж
			monthlyRate := credit.InterestRate / 12 / 100
			if monthlyRate > 0 {
				// Упрощенный расчет для демонстрации
				monthlyPayments += credit.Amount * monthlyRate
			} else {
				monthlyPayments += credit.Amount / 12
			}
		}
		if credit.Status == models.CreditOverdue {
			overdueAmount += credit.Amount
		}
	}

	// Расчет рекомендованного лимита (40% от среднего дохода за 3 месяца)
	recommendedLimit := s.calculateRecommendedLimit(ctx, userID)

	creditUtilization := 0.0
	if totalAmount > 0 {
		creditUtilization = (monthlyPayments / s.getMonthlyIncome(ctx, userID)) * 100
	}

	return &models.CreditBurden{
		TotalCreditAmount: totalAmount,
		MonthlyPayments:   monthlyPayments,
		OverdueAmount:     overdueAmount,
		RecommendedLimit:  recommendedLimit,
		CreditUtilization: creditUtilization,
	}, nil
}

// PredictBalance прогнозирует баланс на N дней
func (s *AnalyticsService) PredictBalance(ctx context.Context, accountID uuid.UUID, days int) (*models.BalancePrediction, error) {
	// Получаем текущий баланс
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Получаем историю за последние 30 дней
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	income, expense, err := s.transactionRepo.GetIncomeExpenseByPeriod(ctx, accountID, startDate, endDate)
	if err != nil {
		income, expense = 0, 0
	}

	// Рассчитываем средние дневные показатели
	avgDailyIncome := income / 30
	avgDailyExpense := expense / 30

	// Прогнозируем баланс
	predictedChange := (avgDailyIncome - avgDailyExpense) * float64(days)
	predictedBalance := account.Balance + predictedChange

	// Рассчитываем уверенность (чем больше данных, тем выше уверенность)
	confidence := 0.5 + (float64(30)/365)*0.5
	if confidence > 0.95 {
		confidence = 0.95
	}

	return &models.BalancePrediction{
		CurrentBalance:   account.Balance,
		PredictedBalance: predictedBalance,
		Confidence:       confidence,
		Days:             days,
		PredictionDate:   endDate.AddDate(0, 0, days),
		AvgDailyIncome:   avgDailyIncome,
		AvgDailyExpense:  avgDailyExpense,
		UpcomingPayments: 0, // Можно добавить расчет из графика платежей
	}, nil
}

// calculateRecommendedLimit рассчитывает рекомендованный кредитный лимит
func (s *AnalyticsService) calculateRecommendedLimit(ctx context.Context, userID uuid.UUID) float64 {
	// 40% от среднего дохода за последние 3 месяца
	accounts, _ := s.accountRepo.FindByUserID(ctx, userID)
	if len(accounts) == 0 {
		return 0
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)
	income, _, _ := s.transactionRepo.GetIncomeExpenseByPeriod(ctx, accounts[0].ID, startDate, endDate)

	avgMonthlyIncome := income / 3
	return avgMonthlyIncome * 0.4
}

// getMonthlyIncome получает средний месячный доход
func (s *AnalyticsService) getMonthlyIncome(ctx context.Context, userID uuid.UUID) float64 {
	accounts, _ := s.accountRepo.FindByUserID(ctx, userID)
	if len(accounts) == 0 {
		return 1
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)
	income, _, _ := s.transactionRepo.GetIncomeExpenseByPeriod(ctx, accounts[0].ID, startDate, endDate)
	return income / 3
}
