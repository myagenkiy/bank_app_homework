package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PaymentScheduleRepository struct {
	db *sql.DB
}

func NewPaymentScheduleRepository(db *sql.DB) *PaymentScheduleRepository {
	return &PaymentScheduleRepository{db: db}
}

func (r *PaymentScheduleRepository) CreateBatch(ctx context.Context, schedules []*models.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (id, credit_id, due_date, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, schedule := range schedules {
		schedule.ID = uuid.New()
		_, err := r.db.ExecContext(ctx, query,
			schedule.ID, schedule.CreditID, schedule.DueDate, schedule.Amount, schedule.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PaymentScheduleRepository) FindByCreditID(ctx context.Context, creditID uuid.UUID) ([]*models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, due_date, amount, status, paid_at FROM payment_schedules WHERE credit_id = $1 ORDER BY due_date ASC`

	rows, err := r.db.QueryContext(ctx, query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		var s models.PaymentSchedule
		err := rows.Scan(&s.ID, &s.CreditID, &s.DueDate, &s.Amount, &s.Status, &s.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *PaymentScheduleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus, paidAt *time.Time) error {
	query := `UPDATE payment_schedules SET status = $2, paid_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, paidAt)
	return err
}

func (r *PaymentScheduleRepository) FindOverduePayments(ctx context.Context, currentDate time.Time) ([]*models.PaymentSchedule, error) {
	query := `
		SELECT ps.id, ps.credit_id, ps.due_date, ps.amount, ps.status, ps.paid_at
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		WHERE ps.due_date < $1 AND ps.status = $2 AND c.status = $3
	`
	rows, err := r.db.QueryContext(ctx, query, currentDate, models.PaymentPending, models.CreditActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		var s models.PaymentSchedule
		err := rows.Scan(&s.ID, &s.CreditID, &s.DueDate, &s.Amount, &s.Status, &s.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, &s)
	}
	return schedules, nil
}
