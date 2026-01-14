# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker-agent ./cmd/agent

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli git

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/worker-agent .

EXPOSE 9090

CMD ["./worker-agent"]
