FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire backend directory
COPY . .

# Change to the gateway service directory
WORKDIR /app/gateway

# Download dependencies (the replace directive should work now)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gateway/main .
COPY --from=builder /app/.env .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"] 