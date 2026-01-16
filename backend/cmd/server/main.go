package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jamesfulreader/gostocks/internal/httpserver"
	"github.com/jamesfulreader/gostocks/internal/stocks"
	"github.com/jamesfulreader/gostocks/pkg/config"
)

func main() {
	// Load environment from .env if present
	_ = config.LoadEnv()

	// Select provider based on env
	alphaKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	finnhubKey := os.Getenv("FINNHUB_API_KEY")

	var provider stocks.Provider

	if alphaKey != "" && finnhubKey != "" {
		alpha := stocks.NewAlphaVantage(alphaKey, nil)
		finnhub := stocks.NewFinnhub(finnhubKey, nil)
		provider = stocks.NewFallback(alpha, finnhub)
		log.Println("Using Fallback provider (Primary: Alpha Vantage, Secondary: Finnhub)")
	} else if alphaKey != "" {
		provider = stocks.NewAlphaVantage(alphaKey, nil)
		log.Println("Using Alpha Vantage provider")
	} else if finnhubKey != "" {
		provider = stocks.NewFinnhub(finnhubKey, nil)
		log.Println("Using Finnhub provider")
	} else {
		provider = stocks.NewMock()
		log.Println("Using Mock provider (set ALPHAVANTAGE_API_KEY and/or FINNHUB_API_KEY to use real data)")
	}

	addr := ":" + config.GetenvDefault("PORT", "8080")

	srv := httpserver.New(provider, addr)
	log.Printf("HTTP server listening on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
