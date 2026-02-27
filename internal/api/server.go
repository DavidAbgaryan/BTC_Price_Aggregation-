package api

import (
	"BTC_Price_Aggregation_/internal/aggregator"
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	agg *aggregator.Service
}

func NewServer(agg *aggregator.Service) *http.ServeMux {
	s := &Server{agg: agg}
	mux := http.NewServeMux()
	mux.HandleFunc("/price", s.handlePrice)
	mux.HandleFunc("/health", s.handleHealth)
	mux.Handle("/metrics", promhttp.Handler())

	return mux

}

func (s *Server) handlePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.agg.GetState())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	state := s.agg.GetState()
	if state.Stale {
		http.Error(w, "All sources failing", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
