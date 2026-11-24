package main

import (
	"fmt"
	"log/slog"
	"os"

	"marku/api/internal/config"
	"marku/api/internal/database"
	"marku/api/internal/logging"
	"marku/api/internal/model"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	slog.SetDefault(logging.New(cfg.Log.Level, cfg.Log.Format))

	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	if err := database.AutoMigrate(db, model.MigrationModels()...); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations completed", "driver", cfg.DB.Driver)
}
