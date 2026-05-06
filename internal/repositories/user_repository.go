package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, username, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	query := `INSERT INTO users (id, email, username, password_hash) VALUES ($1, $2, $3, $4)`
	_, err = r.db.ExecContext(ctx, query, id, email, username, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        id,
		Email:     email,
		Username:  username,
		CreatedAt: time.Now(),
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	query := `SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1`

	var user models.UserWithPassword
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, username, created_at FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CheckPassword(user *models.UserWithPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// GetEmailByUserID получает email пользователя по ID
func (r *UserRepository) GetEmailByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	query := `SELECT email FROM users WHERE id = $1`
	var email string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&email)
	return email, err
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, username, created_at FROM users WHERE id = $1`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
