package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at
	`

	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}

	err := r.db.QueryRowContext(ctx, query,
		transaction.ID,
		transaction.FromAccountID,
		transaction.ToAccountID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.Description,
	).Scan(&transaction.CreatedAt)

	return err
}

func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]*models.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, type, status, description, created_at
		FROM transactions
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID,
			&t.FromAccountID,
			&t.ToAccountID,
			&t.Amount,
			&t.Type,
			&t.Status,
			&t.Description,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}

	return transactions, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error {
	query := `UPDATE transactions SET status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// GetMonthlyStats возвращает статистику по транзакциям за месяц
func (r *TransactionRepository) GetMonthlyStats(ctx context.Context, accountID uuid.UUID, year, month int) (*models.MonthlyStats, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN to_account_id = $1 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN from_account_id = $1 THEN amount ELSE 0 END), 0) as total_expense,
			COUNT(*) as transaction_count
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
			AND status = $2
			AND created_at >= $3
			AND created_at < $4
	`

	stats := &models.MonthlyStats{
		Month: startDate.Format("2006-01"),
	}

	err := r.db.QueryRowContext(ctx, query, accountID, models.StatusCompleted, startDate, endDate).Scan(
		&stats.TotalIncome,
		&stats.TotalExpense,
		&stats.TransactionCount,
	)

	if err != nil {
		return nil, err
	}

	stats.NetChange = stats.TotalIncome - stats.TotalExpense

	return stats, nil
}

// GetIncomeExpenseByPeriod возвращает доходы и расходы за период
func (r *TransactionRepository) GetIncomeExpenseByPeriod(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (income, expense float64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN to_account_id = $1 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN from_account_id = $1 THEN amount ELSE 0 END), 0) as total_expense
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
			AND status = $2
			AND created_at >= $3
			AND created_at < $4
	`

	err = r.db.QueryRowContext(ctx, query, accountID, models.StatusCompleted, startDate, endDate).Scan(&income, &expense)
	return income, expense, err
}
