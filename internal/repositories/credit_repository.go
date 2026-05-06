package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) Create(ctx context.Context, credit *models.Credit) error {
	query := `
		INSERT INTO credits (id, user_id, amount, interest_rate, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	credit.ID = uuid.New()
	_, err := r.db.ExecContext(ctx, query,
		credit.ID, credit.UserID, credit.Amount, credit.InterestRate, credit.Status)
	return err
}

func (r *CreditRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Credit, error) {
	query := `SELECT id, user_id, amount, interest_rate, status, created_at FROM credits WHERE id = $1`

	var credit models.Credit
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&credit.ID, &credit.UserID, &credit.Amount, &credit.InterestRate, &credit.Status, &credit.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &credit, nil
}

func (r *CreditRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Credit, error) {
	query := `SELECT id, user_id, amount, interest_rate, status, created_at FROM credits WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*models.Credit
	for rows.Next() {
		var credit models.Credit
		err := rows.Scan(&credit.ID, &credit.UserID, &credit.Amount, &credit.InterestRate, &credit.Status, &credit.CreatedAt)
		if err != nil {
			return nil, err
		}
		credits = append(credits, &credit)
	}
	return credits, nil
}

func (r *CreditRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CreditStatus) error {
	query := `UPDATE credits SET status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}
