# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o exchange .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/exchange .
COPY --from=builder /app/etc ./etc

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 8888

# Run the application
CMD ["./exchange", "-f", "etc/exchange-api.yaml"]