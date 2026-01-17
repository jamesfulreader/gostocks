package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/jamesfulreader/gostocks/internal/database"
	"github.com/jamesfulreader/gostocks/pkg/config"
)

func TestDatabaseConnection(t *testing.T) {
	// Skip if not running in integration mode or if DB params missing
	// ideally we run this with docker-compose up

	// For this test to run locally, we need env vars.
	// We can try to load from .env in the test or main project root
	_ = config.LoadEnv()

	// Check if we have minimal config to try connecting
	// If we are running in CI or without docker, we might skip
	// But for this task, the goal is verification.

	// We will assume the user (or I) starts the docker container before running this.
	// OR we can start it.

	t.Log("Initializing database connection...")

	// We need to fetch the instance
	// Since New() is a singleton in the current implementation, we just call it.
	dbService := database.New()
	defer dbService.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := dbService.GetPool()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Successfully pinged database!")
}
