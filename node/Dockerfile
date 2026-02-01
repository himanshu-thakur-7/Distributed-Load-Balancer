# ----------Builder Stage--------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY main.go ./

# BUILD static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install curl for healthchecks and debugging
RUN apk add --no-cache curl ca-certificates

# Copy binary from builder
COPY --from=builder /app/backend-server .

# Expose the port the server listens on
EXPOSE 8080

# Run the binary
ENTRYPOINT [ "./backend-server" ]

