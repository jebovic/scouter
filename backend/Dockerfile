# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /scouter-backend ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /scouter-seed ./cmd/seed

# Run stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /scouter-backend .
COPY --from=builder /scouter-seed .

EXPOSE 8080
ENTRYPOINT ["/app/scouter-backend"]
