package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type StockPrice struct {
	ID        int       `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

type Symbol struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

func (s *service) UpsertSymbol(ctx context.Context, sym Symbol) error {
	query := `
		INSERT INTO symbols (symbol, name, type, currency)
		VALUES (@symbol, @name, @type, @currency)
		ON CONFLICT (symbol) DO UPDATE 
		SET name = @name, type = @type, currency = @currency;
	`
	args := pgx.NamedArgs{
		"symbol":   sym.Symbol,
		"name":     sym.Name,
		"type":     sym.Type,
		"currency": sym.Currency,
	}
	_, err := s.pool.Exec(ctx, query, args)
	return err
}

func (s *service) InsertStockPrice(ctx context.Context, price StockPrice) error {
	query := `
		INSERT INTO stock_prices (symbol, price, timestamp)
		VALUES (@symbol, @price, @timestamp)
	`
	args := pgx.NamedArgs{
		"symbol":    price.Symbol,
		"price":     price.Price,
		"timestamp": price.Timestamp,
	}
	_, err := s.pool.Exec(ctx, query, args)
	return err
}

func (s *service) GetLatestStockPrice(ctx context.Context, symbol string) (*StockPrice, error) {
	query := `
		SELECT id, symbol, price, timestamp, created_at
		FROM stock_prices
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	row := s.pool.QueryRow(ctx, query, symbol)
	var p StockPrice
	err := row.Scan(&p.ID, &p.Symbol, &p.Price, &p.Timestamp, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Return nil if no data found
		}
		return nil, fmt.Errorf("failed to scan stock price: %w", err)
	}
	return &p, nil
}

func (s *service) GetAveragePrice(ctx context.Context, symbol string, since time.Time) (float64, error) {
	query := `
		SELECT AVG(price)
		FROM stock_prices
		WHERE symbol = $1 AND timestamp >= $2
	`
	var avg *float64 // Use pointer to handle NULL (no rows)
	err := s.pool.QueryRow(ctx, query, symbol, since).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate average price: %w", err)
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}
