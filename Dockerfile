# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git and ca-certificates (needed for go mod download and swag)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install swag CLI tool
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation
# -g cmd/api/main.go specifies the file containing general API info comments
# --output docs specifies the output directory for the generated files
RUN swag init -g cmd/api/main.go --output docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy the generated swagger docs directory from the builder stage
COPY --from=builder /app/docs ./docs

# Copy the .env file from the build context (root of the project)
COPY .env .env

# Create logs directory and set ownership
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check (ensure your /health endpoint is correctly implemented in your Go app)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
