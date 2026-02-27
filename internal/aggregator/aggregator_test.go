package aggregator

import (
	"BTC_Price_Aggregation_/internal/config"
	"BTC_Price_Aggregation_/internal/provider"
	"context"
	"errors"
	"testing"
	"time"
)

// mockProvider implements the provider.PriceProvider interface for testing.
type mockProvider struct {
	name  string
	price float64
	err   error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) FetchPrice(ctx context.Context) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.price, nil
}

func TestAggregator_PollAndMedianLogic(t *testing.T) {
	// Fast configuration for testing so we don't wait for real timeouts
	cfg := &config.Config{
		PollInterval:   1 * time.Second,
		RequestTimeout: 1 * time.Second,
		MaxRetries:     0, // Disable retries to speed up failure tests
	}

	tests := []struct {
		name            string
		providers       []provider.PriceProvider
		expectedPrice   float64
		expectedSources int
		expectedStale   bool
	}{
		{
			name: "All sources succeed (Odd number - strict median)",
			providers: []provider.PriceProvider{
				&mockProvider{name: "P1", price: 60000.0},
				&mockProvider{name: "P2", price: 61000.0},
				&mockProvider{name: "P3", price: 62000.0},
			},
			expectedPrice:   61000.0,
			expectedSources: 3,
			expectedStale:   false,
		},
		{
			name: "All sources succeed (Even number - average of middle two)",
			providers: []provider.PriceProvider{
				&mockProvider{name: "P1", price: 60000.0},
				&mockProvider{name: "P2", price: 61000.0},
				&mockProvider{name: "P3", price: 62000.0},
				&mockProvider{name: "P4", price: 63000.0},
			},
			expectedPrice:   61500.0, // (61000 + 62000) / 2
			expectedSources: 4,
			expectedStale:   false,
		},
		{
			name: "Partial failure (Service remains resilient)",
			providers: []provider.PriceProvider{
				&mockProvider{name: "P1", price: 60000.0},
				&mockProvider{name: "P2", err: errors.New("API timeout")},
				&mockProvider{name: "P3", price: 62000.0},
			},
			expectedPrice:   61000.0, // Averages the two remaining: (60k + 62k) / 2
			expectedSources: 2,
			expectedStale:   false,
		},
		{
			name: "Total failure (Returns stale flag)",
			providers: []provider.PriceProvider{
				&mockProvider{name: "P1", err: errors.New("500 Internal Error")},
				&mockProvider{name: "P2", err: errors.New("connection reset")},
			},
			expectedPrice:   0, // Initial state price
			expectedSources: 0,
			expectedStale:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			svc := NewService(cfg, tc.providers)
			ctx := context.Background()

			// Act
			// We call poll() directly instead of Start() to run exactly one synchronous cycle
			svc.poll(ctx)

			// Assert
			state := svc.GetState()

			if state.Price != tc.expectedPrice {
				t.Errorf("expected price %v, got %v", tc.expectedPrice, state.Price)
			}

			if state.SourcesUsed != tc.expectedSources {
				t.Errorf("expected sources used %d, got %d", tc.expectedSources, state.SourcesUsed)
			}

			if state.Stale != tc.expectedStale {
				t.Errorf("expected stale %v, got %v", tc.expectedStale, state.Stale)
			}
		})
	}
}
