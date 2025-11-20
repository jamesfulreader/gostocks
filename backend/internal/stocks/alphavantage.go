package stocks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Quote struct {
	Symbol        string   `json:"symbol"`
	Price         float64  `json:"price"`
	Open          float64  `json:"open"`
	High          float64  `json:"high"`
	Low           float64  `json:"low"`
	PreviousClose float64  `json:"previousClose"`
	Change        float64  `json:"change"`
	ChangePercent float64  `json:"changePercent"`
	Timestamp     *string  `json:"timestamp,omitempty"`
}

type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

type Provider interface {
	Quote(ctx context.Context, symbol string) (*Quote, error)
	Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error)
}

type AlphaVantage struct {
	apiKey string
	http   *http.Client
}

func NewAlphaVantage(apiKey string, httpClient *http.Client) *AlphaVantage {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &AlphaVantage{apiKey: apiKey, http: httpClient}
}

func (a *AlphaVantage) Quote(ctx context.Context, symbol string) (*Quote, error) {
	q := url.Values{
		"function": {"GLOBAL_QUOTE"},
		"symbol":   {symbol},
		"apikey":   {a.apiKey},
	}
	u := "https://www.alphavantage.co/query?" + q.Encode()
	resp, err := a.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alphavantage status %d: %s", resp.StatusCode, string(b))
	}
	var raw map[string]map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	m := raw["Global Quote"]
	if len(m) == 0 {
		return nil, errors.New("empty quote")
	}
	parseF := func(k string) float64 { var f float64; _ = json.Unmarshal([]byte(m[k]), &f); if f==0 { fmt.Sscanf(m[k], "%f", &f) }; return f }
	parseS := func(k string) string { return m[k] }
	qp := &Quote{
		Symbol:        parseS("01. symbol"),
		Open:          parseF("02. open"),
		High:          parseF("03. high"),
		Low:           parseF("04. low"),
		Price:         parseF("05. price"),
		PreviousClose: parseF("08. previous close"),
		Change:        parseF("09. change"),
		ChangePercent: parseF("10. change percent"),
	}
	if t := parseS("07. latest trading day"); t != "" {
		qp.Timestamp = &t
	}
	return qp, nil
}

func (a *AlphaVantage) Intraday(ctx context.Context, symbol, interval string, limit int) ([]Candle, error) {
	q := url.Values{
		"function": {"TIME_SERIES_INTRADAY"},
		"symbol":   {symbol},
		"interval": {interval},
		"outputsize": {"compact"},
		"datatype": {"json"},
		"apikey":   {a.apiKey},
	}
	u := "https://www.alphavantage.co/query?" + q.Encode()
	resp, err := a.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alphavantage status %d: %s", resp.StatusCode, string(b))
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	var series map[string]map[string]string
	for k, v := range raw {
		if strings.HasPrefix(k, "Time Series (") {
			if m, ok := v.(map[string]any); ok {
				series = make(map[string]map[string]string, len(m))
				for ts, vv := range m {
					if mm, ok := vv.(map[string]any); ok {
						inner := map[string]string{}
						for kk, vvv := range mm { inner[kk] = fmt.Sprint(vvv) }
						series[ts] = inner
					}
				}
			}
			break
		}
	}
	if series == nil {
		return nil, errors.New("no series")
	}
	candles := make([]Candle, 0, len(series))
	for ts, m := range series {
		var t time.Time
		if tt, err := time.Parse("2006-01-02 15:04:05", ts); err == nil { t = tt }
		pf := func(k string) float64 { var f float64; fmt.Sscanf(m[k], "%f", &f); return f }
		pv := func(k string) int64 { var x int64; fmt.Sscanf(m[k], "%d", &x); return x }
		candles = append(candles, Candle{Time: t, Open: pf("1. open"), High: pf("2. high"), Low: pf("3. low"), Close: pf("4. close"), Volume: pv("5. volume")})
	}
	sort.Slice(candles, func(i, j int) bool { return candles[i].Time.Before(candles[j].Time) })
	if limit > 0 && len(candles) > limit { candles = candles[len(candles)-limit:] }
	return candles, nil
}
