FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git ca-certificates curl

# Set Go environment variables
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies with retry mechanism
RUN for i in 1 2 3 4 5; do \
        echo "=== Attempt $i to download dependencies ===" && \
        go mod download -x && echo "Dependencies downloaded successfully!" && break || \
        (echo "Attempt $i failed, cleaning cache and retrying..." && \
         go clean -modcache && \
         sleep $((i * 3))); \
    done

# Copy source code
COPY . .

# Build application
RUN go build -a -installsuffix cgo -ldflags '-w -s' -o main cmd/server/main.go

# Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and config files
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# Create log directory
RUN mkdir -p logs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Start application
CMD ["./main"]
