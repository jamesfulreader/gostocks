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
	apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	var provider stocks.Provider
	if apiKey != "" {
		provider = stocks.NewAlphaVantage(apiKey, nil)
		log.Println("Using Alpha Vantage provider")
	} else {
		provider = stocks.NewMock()
		log.Println("Using Mock provider (set ALPHAVANTAGE_API_KEY to use Alpha Vantage)")
	}

	addr := ":" + config.GetenvDefault("PORT", "8080")

	srv := httpserver.New(provider, addr)
	log.Printf("HTTP server listening on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
