package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(ctx context.Context, card *models.Card) error {
	query := `
		INSERT INTO cards (id, account_id, encrypted_number, expiry_hash, cvv_hash, hmac, last4, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	card.ID = uuid.New()
	_, err := r.db.ExecContext(ctx, query,
		card.ID, card.AccountID, card.EncryptedNumber,
		card.ExpiryHash, card.CVVHash, card.HMAC, card.Last4)
	return err
}

func (r *CardRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Card, error) {
	query := `SELECT id, account_id, encrypted_number, expiry_hash, cvv_hash, hmac, last4, created_at 
	          FROM cards WHERE account_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.AccountID, &card.EncryptedNumber,
			&card.ExpiryHash, &card.CVVHash, &card.HMAC, &card.Last4, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (r *CardRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Card, error) {
	query := `SELECT id, account_id, encrypted_number, expiry_hash, cvv_hash, hmac, last4, created_at 
	          FROM cards WHERE id = $1`

	var card models.Card
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&card.ID, &card.AccountID, &card.EncryptedNumber,
		&card.ExpiryHash, &card.CVVHash, &card.HMAC, &card.Last4, &card.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &card, nil
}
