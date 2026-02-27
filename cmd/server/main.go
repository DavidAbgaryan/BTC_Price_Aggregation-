package main

import (
	"BTC_Price_Aggregation_/internal/aggregator"
	"BTC_Price_Aggregation_/internal/api"
	"BTC_Price_Aggregation_/internal/config"
	"BTC_Price_Aggregation_/internal/provider"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	cfg := config.Load()

	httpClient := &http.Client{
		Timeout: 5 * time.Second, // Global fallback timeout
	}

	providers := []provider.PriceProvider{
		&provider.Coinbase{HTTPClient: httpClient},
		&provider.Kraken{HTTPClient: httpClient},
	}

	aggService := aggregator.NewService(cfg, providers)
	mux := api.NewServer(aggService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background poller
	go aggService.Start(ctx)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		slog.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
