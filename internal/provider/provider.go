package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// PriceProvider defines the abstraction for any data source
type PriceProvider interface {
	Name() string
	FetchPrice(ctx context.Context) (float64, error)
}

// Coinbase Provider
type Coinbase struct{ HTTPClient *http.Client }

func (c *Coinbase) Name() string { return "Coinbase" }
func (c *Coinbase) FetchPrice(ctx context.Context) (float64, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.coinbase.com/v2/prices/BTC-USD/spot", nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coinbase returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Amount string `json:"amount"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(result.Data.Amount, 64)
}

// Kraken Provider
type Kraken struct{ HTTPClient *http.Client }

func (k *Kraken) Name() string { return "Kraken" }
func (k *Kraken) FetchPrice(ctx context.Context) (float64, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.kraken.com/0/public/Ticker?pair=XXBTZUSD", nil)
	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("kraken returned status: %d", resp.StatusCode)
	}

	var result struct {
		Result map[string]struct {
			C []string `json:"c"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	val := result.Result["XXBTZUSD"].C[0]
	return strconv.ParseFloat(val, 64)
}
