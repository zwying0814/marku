package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"marku/api/internal/config"
	"marku/api/internal/database"
	"marku/api/internal/logging"
	"marku/api/internal/model"
	"marku/api/internal/repository"
	"marku/api/internal/router"
	"marku/api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	slog.SetDefault(logging.New(cfg.Log.Level, cfg.Log.Format))
	if cfg.Auth.JWTSecret == "" && !cfg.IsProduction() {
		slog.Warn("MARKU_JWT_SECRET is empty; set it before enabling admin authentication")
	}

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

	if cfg.DB.AutoMigrate {
		if err := database.AutoMigrate(db, model.MigrationModels()...); err != nil {
			slog.Error("run migrations", "error", err)
			os.Exit(1)
		}
	}

	store := repository.New(db)
	services := service.New(store, cfg)
	created, err := services.SeedAdmin()
	if err != nil {
		slog.Error("seed initial admin", "error", err)
		os.Exit(1)
	}
	if created {
		slog.Warn("default admin created; change the password immediately after login",
			"username", service.DefaultAdminUsername)
	}

	engine, err := router.New(cfg, db, services)
	if err != nil {
		slog.Error("build router", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           engine,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info(
			"marku api listening",
			"addr", cfg.HTTP.Addr,
			"env", cfg.Env,
			"db_driver", cfg.DB.Driver,
			"auto_migrate", cfg.DB.AutoMigrate,
		)
		serverErrors <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case sig := <-stop:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("marku api stopped")
}
