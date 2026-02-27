# Resilient git init

A production-ready Go microservice designed to fetch, aggregate, and serve the current price of Bitcoin (BTC) in USD.

This service demonstrates clean architecture, concurrency safety, and robust error handling. It polls multiple public cryptocurrency APIs (Coinbase and Kraken), calculates a deterministic aggregated price (Median), and exposes the data via a fast HTTP HTTP API.

## 🚀 Key Features
* **Multi-Source Polling:** Concurrently fetches data from independent exchanges.
* **Deterministic Aggregation:** Uses the median price of all healthy sources to prevent outliers from skewing the data.
* **Resilience & Fault Tolerance:** Implements context-aware timeouts and exponential backoff retries. The service remains operational even if partial sources fail.
* **Thread Safety:** Safely manages background polling and concurrent API requests using read-write mutexes (`sync.RWMutex`).
* **Observability:** Provides structured logging (`slog`) and exposes Prometheus-compatible metrics.
* **Graceful Shutdown:** Safely drains connections and stops background workers on OS interruption signals.

---

## 🛠 Prerequisites
* [Docker](https://docs.docker.com/get-docker/)
* Go 1.22+ (Only if running tests locally without Docker)

---

## 📦 Getting Started

### first step
go mod tidy
### Testing
go test -v -cover ./...
###

### Option 1: Using Docker


1. **Build and start the service in the background:**
    docker build --no-cache -t btc-price-aggregation .
    docker run -d -p 8000:8080 --name btc-service btc-price-aggregation

    for production ready
    go mod docker run -d \
      -p 8000:8080 \
      -e POLL_INTERVAL=5s \
      -e MAX_RETRIES=5 \
      -e REQUEST_TIMEOUT=2s \
      --name btc-service \
      btc-price-aggregation



endpoints
http://localhost:8000/price
http://localhost:8000/health
http://localhost:8000/metrics