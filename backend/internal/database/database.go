package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	Health() map[string]string
	Close()
	GetPool() *pgxpool.Pool
}

type service struct {
	pool *pgxpool.Pool
}

var dbInstance *service

func New() Service {
	// reuse instance
	if dbInstance != nil {
		return dbInstance
	}

	// Default to docker-compose service name "db"
	dbHost := "db"
	if host := os.Getenv("DB_HOST"); host != "" {
		dbHost = host
	}

	dbPort := "5432"
	if port := os.Getenv("DB_PORT"); port != "" {
		dbPort = port
	}

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		dbHost,
		dbPort,
		os.Getenv("POSTGRES_DB"),
	)

	log.Printf("Connecting to database at %s:%s", dbHost, dbPort)

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalf("Unable to parse database config: %v\n", err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	dbInstance = &service{
		pool: db,
	}
	return dbInstance
}

func (s *service) GetPool() *pgxpool.Pool {
	return s.pool
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.pool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log fatal for now as we want to know
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"
	stats["open_connections"] = fmt.Sprintf("%d", s.pool.Stat().TotalConns())
	return stats
}

func (s *service) Close() {
	log.Printf("Disconnected from database: %s", os.Getenv("POSTGRES_DB"))
	s.pool.Close()
	dbInstance = nil
}
