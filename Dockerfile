# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o btc-price-aggregation ./cmd/server

# Stage 2: Create the final lean image
FROM alpine:latest

# ✅ FIX: Install CA certificates so Go can make HTTPS requests!
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/btc-price-aggregation .

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./btc-price-aggregation"]