package users

import (
	"context"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	AddToPortfolio(ctx context.Context, userID int, symbol string) error
	GetPortfolio(ctx context.Context, userID int) ([]string, error)
	RemoveFromPortfolio(ctx context.Context, userID int, symbol string) error
}
