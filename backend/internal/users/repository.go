package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at",
		email, passwordHash).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.PasswordHash = passwordHash
	return &user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, "SELECT id, email, password_hash, created_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) AddToPortfolio(ctx context.Context, userID int, symbol string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO user_portfolios (user_id, symbol) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, symbol)
	if err != nil {
		return fmt.Errorf("failed to add to portfolio: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetPortfolio(ctx context.Context, userID int) ([]string, error) {
	rows, err := r.db.Query(ctx, "SELECT symbol FROM user_portfolios WHERE user_id = $1 ORDER BY added_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

func (r *PostgresRepository) RemoveFromPortfolio(ctx context.Context, userID int, symbol string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM user_portfolios WHERE user_id = $1 AND symbol = $2", userID, symbol)
	if err != nil {
		return fmt.Errorf("failed to remove from portfolio: %w", err)
	}
	return nil
}
