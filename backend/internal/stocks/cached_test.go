package stocks_test

import (
	"context"
	"testing"
	"time"

	"github.com/jamesfulreader/gostocks/internal/database"
	"github.com/jamesfulreader/gostocks/internal/stocks"
	// Assuming we can use gomock or just manual mocks
)

// Manual mocks to avoid generating code
type MockUpstream struct {
	Count int
}

func (m *MockUpstream) Quote(ctx context.Context, symbol string) (*stocks.Quote, error) {
	m.Count++
	t := time.Now().Format(time.RFC3339)
	return &stocks.Quote{
		Symbol:    symbol,
		Price:     100.0 + float64(m.Count), // Different price each time
		Timestamp: &t,
	}, nil
}
func (m *MockUpstream) Intraday(ctx context.Context, symbol, interval string, limit int) ([]stocks.Candle, error) {
	return nil, nil
}

type MockDB struct {
	Store map[string]database.StockPrice
}

func (m *MockDB) GetLatestStockPrice(ctx context.Context, symbol string) (*database.StockPrice, error) {
	if v, ok := m.Store[symbol]; ok {
		return &v, nil
	}
	return nil, nil // Not found
}
func (m *MockDB) InsertStockPrice(ctx context.Context, price database.StockPrice) error {
	m.Store[price.Symbol] = price
	return nil
}
func (m *MockDB) UpsertSymbol(ctx context.Context, sym database.Symbol) error {
	return nil
}
func (m *MockDB) GetAveragePrice(ctx context.Context, symbol string, since time.Time) (float64, error) {
	return 0, nil
}

func TestCachedProvider(t *testing.T) {
	upstream := &MockUpstream{}
	db := &MockDB{Store: make(map[string]database.StockPrice)}

	// Create provider with 1 second TTL for test
	provider := stocks.NewCachedProvider(upstream, db, 1*time.Second)

	ctx := context.Background()
	symbol := "TEST"

	// 1. First Call - Should hit Upstream
	q1, err := provider.Quote(ctx, symbol)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if upstream.Count != 1 {
		t.Errorf("Expected upstream count 1, got %d", upstream.Count)
	}
	if q1.Price != 101.0 {
		t.Errorf("Expected price 101.0, got %f", q1.Price)
	}

	// Wait a bit for async goroutine to save to DB (if we made it async, but we kept it sync-ish or fast enough?)
	// In the implementation it IS async: go func(val *Quote) { ... }
	// So we need to wait / poll.
	time.Sleep(100 * time.Millisecond)

	// 2. Second Call (Immediate) - Should hit Cache
	q2, err := provider.Quote(ctx, symbol)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	// Upstream count should STILL be 1
	if upstream.Count != 1 {
		t.Errorf("Expected upstream count 1 (Cache Hit), got %d", upstream.Count)
	}
	if q2.Price != 101.0 {
		t.Errorf("Expected cached price 101.0, got %f", q2.Price)
	}

	// 3. Third Call (After TTL) - Should hit Upstream again
	time.Sleep(1100 * time.Millisecond) // Wait > 1s
	q3, err := provider.Quote(ctx, symbol)
	if err != nil {
		t.Fatalf("Third call failed: %v", err)
	}
	if upstream.Count != 2 {
		t.Errorf("Expected upstream count 2 (Cache Expired), got %d", upstream.Count)
	}
	if q3.Price != 102.0 {
		t.Errorf("Expected fresh price 102.0, got %f", q3.Price)
	}
}
