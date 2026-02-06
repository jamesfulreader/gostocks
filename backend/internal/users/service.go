package users

import (
	"context"
	"errors"

	"github.com/jamesfulreader/gostocks/internal/stocks"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo     Repository
	provider stocks.Provider
}

func NewService(repo Repository, provider stocks.Provider) *Service {
	return &Service{
		repo:     repo,
		provider: provider,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateUser(ctx, email, string(hashedPassword))
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (s *Service) GetPortfolio(ctx context.Context, userID int) ([]string, error) {
	return s.repo.GetPortfolio(ctx, userID)
}

func (s *Service) AddToPortfolio(ctx context.Context, userID int, symbol string) error {
	// Validate symbol existence (and side-effect: upsert to DB via CachedProvider)
	q, err := s.provider.Quote(ctx, symbol)
	if err != nil {
		return err
	}
	return s.repo.AddToPortfolio(ctx, userID, q.Symbol)
}

func (s *Service) RemoveFromPortfolio(ctx context.Context, userID int, symbol string) error {
	return s.repo.RemoveFromPortfolio(ctx, userID, symbol)
}
