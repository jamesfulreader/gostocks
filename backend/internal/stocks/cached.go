package stocks

import (
	"context"
	"log"
	"time"

	"github.com/jamesfulreader/gostocks/internal/database"
)

type DatabaseService interface {
	GetLatestStockPrice(ctx context.Context, symbol string) (*database.StockPrice, error)
	InsertStockPrice(ctx context.Context, price database.StockPrice) error
	UpsertSymbol(ctx context.Context, sym database.Symbol) error
}

type CachedProvider struct {
	Upstream Provider
	DB       DatabaseService
	CacheTTL time.Duration
}

func NewCachedProvider(upstream Provider, db DatabaseService, ttl time.Duration) *CachedProvider {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &CachedProvider{
		Upstream: upstream,
		DB:       db,
		CacheTTL: ttl,
	}
}

func (c *CachedProvider) Quote(ctx context.Context, symbol string) (*Quote, error) {
	// 1. Check DB for fresh data
	latest, err := c.DB.GetLatestStockPrice(ctx, symbol)
	if err == nil && latest != nil {
		if time.Since(latest.Timestamp) < c.CacheTTL {
			log.Printf("Create Cache Hit for %s", symbol)
			return &Quote{
				Symbol: latest.Symbol,
				Price:  latest.Price,
				// Timestamp is *string in Quote struct, adapting it
				Timestamp: stringPointer(latest.Timestamp.Format(time.RFC3339)),
			}, nil
		}
	}

	// 2. Call Upstream
	q, err := c.Upstream.Quote(ctx, symbol)
	if err != nil {
		// If upstream fails, maybe return stale data if available?
		if latest != nil {
			log.Printf("Upstream failed, returning stale data for %s", symbol)
			return &Quote{
				Symbol:    latest.Symbol,
				Price:     latest.Price,
				Timestamp: stringPointer(latest.Timestamp.Format(time.RFC3339)),
			}, nil
		}
		return nil, err
	}

	// 3. Save to DB (Async to not block response?) - Synchronous for now for data integrity
	go func(val *Quote) {
		// Create a detached context for the db operation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Ensure symbol exists
		_ = c.DB.UpsertSymbol(ctx, database.Symbol{
			Symbol: val.Symbol,
			Name:   val.Symbol, // We don't have name from Quote, using Symbol as placeholder
			Type:   "Unknown",
		})

		// Save price
		ts := time.Now()
		if val.Timestamp != nil {
			if t, err := time.Parse(time.RFC3339, *val.Timestamp); err == nil {
				ts = t
			}
		}

		err := c.DB.InsertStockPrice(ctx, database.StockPrice{
			Symbol:    val.Symbol,
			Price:     val.Price,
			Timestamp: ts,
		})
		if err != nil {
			log.Printf("Failed to cache price for %s: %v", val.Symbol, err)
		}
	}(q)

	return q, nil
}

func (c *CachedProvider) Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	// For now, passthrough. Caching time-series is more complex.
	return c.Upstream.Intraday(ctx, symbol, interval, limit)
}

func stringPointer(s string) *string {
	return &s
}
