# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install dependencies
RUN npm install -g pnpm && pnpm install

# Copy frontend source
COPY frontend/ .

# Build frontend
RUN pnpm build

# Backend build stage
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from backend builder
COPY --from=backend-builder /app/main .

# Copy frontend build
COPY --from=frontend-builder /app/frontend/dist ./static

# Copy configuration files
COPY prometheus.yml /etc/prometheus/prometheus.yml
COPY grafana-dashboards /etc/grafana/dashboards

# Expose ports
EXPOSE 2112 3000

# Run
CMD ["./main"]
