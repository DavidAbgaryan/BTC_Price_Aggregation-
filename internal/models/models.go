package models

import "time"

type PriceResponse struct {
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	SourcesUsed int       `json:"sources_used"`
	LastUpdated time.Time `json:"last_updated"`
	Stale       bool      `json:"stale"`
}
