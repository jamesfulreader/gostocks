package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jamesfulreader/gostocks/internal/database"
	"github.com/jamesfulreader/gostocks/internal/httpserver"
	"github.com/jamesfulreader/gostocks/internal/stocks"
	"github.com/jamesfulreader/gostocks/pkg/config"
)

func main() {
	// Load environment from .env if present
	_ = config.LoadEnv()

	// Initialize Database
	db := database.New()
	defer db.Close()

	// Check DB Health
	health := db.Health()
	log.Printf("Database health: %v", health)

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

	// Wrap with Caching Provider
	// Use a 5-minute TTL
	provider = stocks.NewCachedProvider(provider, db, 5*time.Minute)
	log.Println("Enabled Database Caching for Stock Provider")

	addr := ":" + config.GetenvDefault("PORT", "8080")

	srv := httpserver.New(provider, db.GetPool(), addr)
	log.Printf("HTTP server listening on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
