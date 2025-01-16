# Build stage
FROM golang:1.23.4 AS builder

# Set the working directory
WORKDIR /build

# Copy source code
COPY ./ ./

# Build the application
WORKDIR /build/main
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o seniorlabai .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and required files from builder
COPY --from=builder /build/main/seniorlabai .
COPY --from=builder /build/.env .
COPY --from=builder /build/main/templates ./templates

# Create directory for log files with proper permissions
RUN mkdir -p /app/logs && \
    chmod 755 /app/logs && \
    touch /app/logs/logfile.log /app/logs/usage.log && \
    chmod 666 /app/logs/logfile.log /app/logs/usage.log

# Expose the necessary port
EXPOSE 8468

# Set the entrypoint
ENTRYPOINT ["/app/seniorlabai"]
