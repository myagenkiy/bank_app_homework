package services

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
)

type TransferService struct {
	db              *sql.DB
	accountRepo     *repositories.AccountRepository
	transactionRepo *repositories.TransactionRepository
}

func NewTransferService(
	db *sql.DB,
	accountRepo *repositories.AccountRepository,
	transactionRepo *repositories.TransactionRepository,
) *TransferService {
	return &TransferService{
		db:              db,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

// Deposit - пополнение счета
func (s *TransferService) Deposit(ctx context.Context, req *models.DepositRequest) (*models.Transaction, error) {
	log.Println("Deposit started for account:", req.AccountID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return nil, err
	}
	defer tx.Rollback()

	var account models.Account
	query := `SELECT id, user_id, balance, currency FROM accounts WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, req.AccountID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Currency)
	if err != nil {
		log.Println("Account not found:", err)
		return nil, errors.New("account not found")
	}

	transaction := &models.Transaction{
		ID:          uuid.New(),
		ToAccountID: &req.AccountID,
		Amount:      req.Amount,
		Type:        models.DepositTransaction,
		Status:      models.StatusPending,
		Description: req.Description,
	}

	insertQuery := `
		INSERT INTO transactions (id, to_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		transaction.ID, transaction.ToAccountID, transaction.Amount,
		transaction.Type, transaction.Status, transaction.Description)
	if err != nil {
		log.Println("Failed to insert transaction:", err)
		return nil, err
	}

	newBalance := account.Balance + req.Amount
	updateQuery := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, req.AccountID, newBalance)
	if err != nil {
		log.Println("Failed to update balance:", err)
		return nil, err
	}

	transaction.Status = models.StatusCompleted
	updateStatusQuery := `UPDATE transactions SET status = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateStatusQuery, transaction.ID, transaction.Status)
	if err != nil {
		log.Println("Failed to update transaction status:", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Failed to commit transaction:", err)
		return nil, err
	}

	log.Println("Deposit completed successfully")
	return transaction, nil
}

// Withdraw - снятие денег
func (s *TransferService) Withdraw(ctx context.Context, req *models.WithdrawRequest) (*models.Transaction, error) {
	log.Println("Withdraw started for account:", req.AccountID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var account models.Account
	query := `SELECT id, user_id, balance, currency FROM accounts WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, req.AccountID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Currency)
	if err != nil {
		return nil, errors.New("account not found")
	}

	if account.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	transaction := &models.Transaction{
		ID:            uuid.New(),
		FromAccountID: &req.AccountID,
		Amount:        req.Amount,
		Type:          models.WithdrawTransaction,
		Status:        models.StatusPending,
		Description:   req.Description,
	}

	insertQuery := `
		INSERT INTO transactions (id, from_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		transaction.ID, transaction.FromAccountID, transaction.Amount,
		transaction.Type, transaction.Status, transaction.Description)
	if err != nil {
		return nil, err
	}

	newBalance := account.Balance - req.Amount
	updateQuery := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, req.AccountID, newBalance)
	if err != nil {
		return nil, err
	}

	transaction.Status = models.StatusCompleted
	updateStatusQuery := `UPDATE transactions SET status = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateStatusQuery, transaction.ID, transaction.Status)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Transfer - перевод между счетами
func (s *TransferService) Transfer(ctx context.Context, req *models.TransferRequest) (*models.Transaction, error) {
	log.Println("Transfer started from:", req.FromAccountID, "to:", req.ToAccountID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var fromAccount, toAccount models.Account

	query := `SELECT id, user_id, balance, currency FROM accounts WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, req.FromAccountID).Scan(
		&fromAccount.ID, &fromAccount.UserID, &fromAccount.Balance, &fromAccount.Currency)
	if err != nil {
		return nil, errors.New("from account not found")
	}

	err = tx.QueryRowContext(ctx, query, req.ToAccountID).Scan(
		&toAccount.ID, &toAccount.UserID, &toAccount.Balance, &toAccount.Currency)
	if err != nil {
		return nil, errors.New("to account not found")
	}

	if fromAccount.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	transaction := &models.Transaction{
		ID:            uuid.New(),
		FromAccountID: &req.FromAccountID,
		ToAccountID:   &req.ToAccountID,
		Amount:        req.Amount,
		Type:          models.TransferTransaction,
		Status:        models.StatusPending,
		Description:   req.Description,
	}

	insertQuery := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		transaction.ID, transaction.FromAccountID, transaction.ToAccountID,
		transaction.Amount, transaction.Type, transaction.Status, transaction.Description)
	if err != nil {
		return nil, err
	}

	newFromBalance := fromAccount.Balance - req.Amount
	newToBalance := toAccount.Balance + req.Amount

	updateQuery := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, req.FromAccountID, newFromBalance)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, updateQuery, req.ToAccountID, newToBalance)
	if err != nil {
		return nil, err
	}

	transaction.Status = models.StatusCompleted
	updateStatusQuery := `UPDATE transactions SET status = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateStatusQuery, transaction.ID, transaction.Status)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// ProcessScheduledPayment - списание запланированного платежа (для шедулера)
func (s *TransferService) ProcessScheduledPayment(ctx context.Context, accountID uuid.UUID, amount float64, description string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var account models.Account
	query := `SELECT id, user_id, balance, currency FROM accounts WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, accountID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Currency)
	if err != nil {
		return errors.New("account not found")
	}

	if account.Balance < amount {
		return errors.New("insufficient funds for scheduled payment")
	}

	transaction := &models.Transaction{
		ID:            uuid.New(),
		FromAccountID: &accountID,
		Amount:        amount,
		Type:          models.WithdrawTransaction,
		Status:        models.StatusPending,
		Description:   description,
	}

	insertQuery := `
		INSERT INTO transactions (id, from_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		transaction.ID, transaction.FromAccountID, transaction.Amount,
		transaction.Type, transaction.Status, transaction.Description)
	if err != nil {
		return err
	}

	newBalance := account.Balance - amount
	updateQuery := `UPDATE accounts SET balance = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, accountID, newBalance)
	if err != nil {
		return err
	}

	transaction.Status = models.StatusCompleted
	updateStatusQuery := `UPDATE transactions SET status = $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateStatusQuery, transaction.ID, transaction.Status)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *TransferService) GetTransactionHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]*models.Transaction, error) {
	return s.transactionRepo.GetByAccountID(ctx, accountID, limit)
}

func (s *TransferService) GetAccount(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	return s.accountRepo.FindByID(ctx, accountID)
}
