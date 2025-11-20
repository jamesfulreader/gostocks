package stocks

import (
	"context"
	"time"
)

type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Quote(ctx context.Context, symbol string) (*Quote, error) {
	// Return a deterministic mock quote
	now := time.Now().Format("2006-01-02")
	return &Quote{
		Symbol:        symbol,
		Price:         123.45,
		Open:          120.00,
		High:          125.00,
		Low:           119.50,
		PreviousClose: 121.00,
		Change:        2.45,
		ChangePercent: 2.02,
		Timestamp:     &now,
	}, nil
}

func (m *Mock) Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	now := time.Now().Truncate(time.Minute)
	if limit <= 0 { limit = 60 }
	out := make([]Candle, 0, limit)
	base := 120.0
	for i := limit - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * time.Minute)
		open := base + float64(i%5)
		close := open + 0.5
		high := close + 0.25
		low := open - 0.25
		vol := int64(1000 + i*3)
		out = append(out, Candle{Time: t, Open: open, High: high, Low: low, Close: close, Volume: vol})
	}
	return out, nil
}
