package aggregator

import (
	"BTC_Price_Aggregation_/internal/config"
	"BTC_Price_Aggregation_/internal/models"
	"BTC_Price_Aggregation_/internal/provider"
	"BTC_Price_Aggregation_/pkg/retry"
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	fetchSuccess = promauto.NewCounterVec(prometheus.CounterOpts{Name: "fetch_success_total"}, []string{"source"})
	fetchFailure = promauto.NewCounterVec(prometheus.CounterOpts{Name: "fetch_failure_total"}, []string{"source"})
	currentPrice = promauto.NewGauge(prometheus.GaugeOpts{Name: "current_btc_price"})
)

type Service struct {
	providers []provider.PriceProvider
	config    *config.Config

	mu    sync.RWMutex
	state models.PriceResponse
}

func NewService(cfg *config.Config, providers []provider.PriceProvider) *Service {
	return &Service{
		providers: providers,
		config:    cfg,
		state: models.PriceResponse{
			Stale: true, // Initially stale until first successful poll
		},
	}
}

// Start runs the background poller
func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	slog.Info("Starting aggregator poller", "interval", s.config.PollInterval)
	s.poll(ctx) // initial poll

	for {
		select {
		case <-ticker.C:
			s.poll(ctx)
		case <-ctx.Done():
			slog.Info("Stopping aggregator poller")
			return
		}
	}
}

func (s *Service) poll(ctx context.Context) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var prices []float64

	for _, p := range s.providers {
		wg.Add(1)
		go func(prov provider.PriceProvider) {
			defer wg.Done()

			var price float64
			// Implement Retry and Timeout per request
			err := retry.Do(ctx, s.config.MaxRetries, func() error {
				reqCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
				defer cancel()

				start := time.Now()
				var fetchErr error
				price, fetchErr = prov.FetchPrice(reqCtx)

				if fetchErr != nil {
					slog.Warn("Fetch failed, retrying...", "source", prov.Name(), "error", fetchErr)
					return fetchErr
				}
				slog.Debug("Fetch success", "source", prov.Name(), "latency", time.Since(start))
				return nil
			})

			if err != nil {
				fetchFailure.WithLabelValues(prov.Name()).Inc()
				slog.Error("Source completely failed", "source", prov.Name(), "error", err)
				return
			}

			fetchSuccess.WithLabelValues(prov.Name()).Inc()
			mu.Lock()
			prices = append(prices, price)
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	s.updateState(prices)
}

func (s *Service) updateState(prices []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(prices) == 0 {
		s.state.Stale = true
		slog.Warn("All sources failed, data is now stale")
		return
	}

	// Calculate Median
	sort.Float64s(prices)
	var median float64
	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		median = (prices[mid-1] + prices[mid]) / 2.0
	} else {
		median = prices[mid]
	}

	s.state = models.PriceResponse{
		Price:       median,
		Currency:    "USD",
		SourcesUsed: len(prices),
		LastUpdated: time.Now().UTC(),
		Stale:       false,
	}

	currentPrice.Set(median)
	slog.Info("State updated", "price", median, "sources_used", len(prices))
}

// GetState safely reads the current state
func (s *Service) GetState() models.PriceResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
