package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jamesfulreader/gostocks/internal/database"
)

func TestStockPricePersistence(t *testing.T) {
	// Setup connection
	if os.Getenv("DB_HOST") == "" {
		_ = os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		_ = os.Setenv("DB_PORT", "5432")
	}

	// Assuming these are set in env or defaults work, but better to set explicitly for test if needed
	if os.Getenv("POSTGRES_USER") == "" {
		_ = os.Setenv("POSTGRES_USER", "user")
		_ = os.Setenv("POSTGRES_PASSWORD", "password")
		_ = os.Setenv("POSTGRES_DB", "gostocks")
	}

	t.Log("Connecting to database...")
	db := database.New()
	defer db.Close() // In a real app we might not close, but for test isolation it's good practice provided it doesn't break pool

	ctx := context.Background()

	// 1. Insert a Symbol
	sym := database.Symbol{
		Symbol:   "TEST",
		Name:     "Test Company",
		Type:     "Equity",
		Currency: "USD",
	}
	if err := db.(interface {
		UpsertSymbol(context.Context, database.Symbol) error
	}).UpsertSymbol(ctx, sym); err != nil {
		t.Fatalf("Failed to upsert symbol: %v", err)
	}

	// 2. Insert Old Price Data
	oldPrice := database.StockPrice{
		Symbol:    "TEST",
		Price:     150.00,
		Timestamp: time.Now().Add(-24 * time.Hour), // 1 day ago
	}
	if err := db.(interface {
		InsertStockPrice(context.Context, database.StockPrice) error
	}).InsertStockPrice(ctx, oldPrice); err != nil {
		t.Fatalf("Failed to insert old price: %v", err)
	}

	// 3. Insert Newer Price Data (to verify we can fetch latest)
	newPrice := database.StockPrice{
		Symbol:    "TEST",
		Price:     155.50,
		Timestamp: time.Now(),
	}
	if err := db.(interface {
		InsertStockPrice(context.Context, database.StockPrice) error
	}).InsertStockPrice(ctx, newPrice); err != nil {
		t.Fatalf("Failed to insert new price: %v", err)
	}

	// 4. Retrieve Latest Price (Referencing "old" data in DB)
	latest, err := db.(interface {
		GetLatestStockPrice(context.Context, string) (*database.StockPrice, error)
	}).GetLatestStockPrice(ctx, "TEST")
	if err != nil {
		t.Fatalf("Failed to get latest price: %v", err)
	}
	if latest == nil {
		t.Fatal("Expected price data, got nil")
	}

	t.Logf("Retrieved Price: Symbol=%s, Price=%.2f, Time=%s", latest.Symbol, latest.Price, latest.Timestamp)

	if latest.Price != newPrice.Price {
		t.Errorf("Expected price %.2f, got %.2f", newPrice.Price, latest.Price)
	}

	// 5. Test Average Price (Trend)
	// We have 150.00 and 155.50. Average should be 152.75
	since := time.Now().Add(-48 * time.Hour)
	avg, err := db.(interface {
		GetAveragePrice(context.Context, string, time.Time) (float64, error)
	}).GetAveragePrice(ctx, "TEST", since)
	if err != nil {
		t.Fatalf("Failed to get average price: %v", err)
	}
	t.Logf("Average Price (Last 48h): %.2f", avg)

	expectedAvg := (150.00 + 155.50) / 2
	if avg != expectedAvg {
		t.Errorf("Expected average %.2f, got %.2f", expectedAvg, avg)
	}
}
