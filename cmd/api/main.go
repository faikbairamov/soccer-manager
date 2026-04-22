package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/faikbairamov/soccer-manager/internal/config"
	"github.com/faikbairamov/soccer-manager/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	pool, err := db.Connect(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("database connection established")
}
