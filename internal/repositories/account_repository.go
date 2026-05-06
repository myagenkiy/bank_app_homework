package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository создает новый репозиторий для счетов
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create создает новый банковский счет
func (r *AccountRepository) Create(ctx context.Context, userID uuid.UUID, currency string) (*models.Account, error) {
	account := &models.Account{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  0,
		Currency: currency,
	}

	query := `INSERT INTO accounts (id, user_id, balance, currency) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, account.ID, account.UserID, account.Balance, account.Currency)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// FindByUserID находит все счета пользователя
func (r *AccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}

// FindByID находит счет по ID
func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1`

	var account models.Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// UpdateBalance обновляет баланс счета
func (r *AccountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance float64) error {
	query := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, newBalance)
	return err
}

func (r *AccountRepository) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, newBalance float64) error {
	query := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id, newBalance)
	return err
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1 FOR UPDATE`

	var account models.Account
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &account, nil
}
