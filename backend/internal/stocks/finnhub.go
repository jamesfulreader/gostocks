package stocks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type Finnhub struct {
	apiKey string
	http   *http.Client
}

func NewFinnhub(apiKey string, httpClient *http.Client) *Finnhub {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Finnhub{apiKey: apiKey, http: httpClient}
}

// FinnhubQuote matches https://finnhub.io/docs/api/quote
type FinnhubQuote struct {
	C  float64 `json:"c"`  // Current price
	D  float64 `json:"d"`  // Change
	DP float64 `json:"dp"` // Percent change
	H  float64 `json:"h"`  // High
	L  float64 `json:"l"`  // Low
	O  float64 `json:"o"`  // Open
	PC float64 `json:"pc"` // Previous close
	T  int64   `json:"t"`  // Timestamp
}

func (f *Finnhub) Quote(ctx context.Context, symbol string) (*Quote, error) {
	q := url.Values{
		"symbol": {symbol},
		"token":  {f.apiKey},
	}
	u := "https://finnhub.io/api/v1/quote?" + q.Encode()

	resp, err := f.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("finnhub status %d: %s", resp.StatusCode, string(b))
	}

	var raw FinnhubQuote
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	// Finnhub returns 0s if symbol not found, but status 200.
	if raw.C == 0 && raw.O == 0 && raw.PC == 0 {
		return nil, fmt.Errorf("symbol not found or empty data for %s", symbol)
	}

	ts := time.Unix(raw.T, 0).Format(time.RFC3339)

	return &Quote{
		Symbol:        symbol,
		Price:         raw.C,
		Open:          raw.O,
		High:          raw.H,
		Low:           raw.L,
		PreviousClose: raw.PC,
		Change:        raw.D,
		ChangePercent: raw.DP,
		Timestamp:     &ts,
	}, nil
}

// FinnhubCandles matches https://finnhub.io/docs/api/stock-candles
type FinnhubCandles struct {
	C []float64 `json:"c"`
	H []float64 `json:"h"`
	L []float64 `json:"l"`
	O []float64 `json:"o"`
	S string    `json:"s"` // Status
	T []int64   `json:"t"`
	V []float64 `json:"v"` // Volume can be float in JSON sometimes? doc says int/float usually
}

func (f *Finnhub) Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	// Map our interval to Finnhub resolution
	// Supported: 1, 5, 15, 30, 60, D, W, M
	resolution := "D" // Default to Daily as per plan

	now := time.Now()
	// Fetch 1 year of data to ensure we hit the limit
	from := now.AddDate(-1, 0, 0).Unix()
	to := now.Unix()

	q := url.Values{
		"symbol":     {symbol},
		"resolution": {resolution},
		"from":       {fmt.Sprint(from)},
		"to":         {fmt.Sprint(to)},
		"token":      {f.apiKey},
	}
	u := "https://finnhub.io/api/v1/stock/candle?" + q.Encode()

	resp, err := f.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("finnhub status %d: %s", resp.StatusCode, string(b))
	}

	var raw FinnhubCandles
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if raw.S != "ok" {
		return nil, fmt.Errorf("finnhub error status: %s", raw.S)
	}

	count := len(raw.T)
	if count == 0 {
		return []Candle{}, nil
	}

	candles := make([]Candle, 0, count)
	for i := 0; i < count; i++ {
		t := time.Unix(raw.T[i], 0)
		candles = append(candles, Candle{
			Time:   t,
			Open:   raw.O[i],
			High:   raw.H[i],
			Low:    raw.L[i],
			Close:  raw.C[i],
			Volume: int64(raw.V[i]),
		})
	}

	// Sort just in case, though usually returned sorted
	sort.Slice(candles, func(i, j int) bool { return candles[i].Time.Before(candles[j].Time) })

	if limit > 0 && len(candles) > limit {
		candles = candles[len(candles)-limit:]
	}

	return candles, nil
}
