FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire backend directory
COPY . .

# Change to the auth service directory
WORKDIR /app/services/auth

# Download dependencies (the replace directive should work now)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/auth/main .

# Expose port
EXPOSE 8081

# Run the binary
CMD ["./main"] 