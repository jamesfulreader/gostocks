package stocks

import (
	"context"
	"log"
)

// Fallback provider wraps two providers: Primary and Secondary.
// It attempts to call Primary first. If Primary returns an error,
// it logs the error and retries with Secondary.
type Fallback struct {
	Primary   Provider
	Secondary Provider
}

func NewFallback(primary, secondary Provider) *Fallback {
	return &Fallback{
		Primary:   primary,
		Secondary: secondary,
	}
}

func (f *Fallback) Quote(ctx context.Context, symbol string) (*Quote, error) {
	q, err := f.Primary.Quote(ctx, symbol)
	if err == nil {
		return q, nil
	}

	log.Printf("Primary provider failed for Quote(%s): %v. Switching to secondary.", symbol, err)
	return f.Secondary.Quote(ctx, symbol)
}

func (f *Fallback) Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	c, err := f.Primary.Intraday(ctx, symbol, interval, limit)
	if err == nil {
		return c, nil
	}

	log.Printf("Primary provider failed for Intraday(%s): %v. Switching to secondary.", symbol, err)
	return f.Secondary.Intraday(ctx, symbol, interval, limit)
}
