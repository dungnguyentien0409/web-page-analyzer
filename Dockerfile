FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o bin/analyzer \
    ./cmd/server/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/analyzer .
COPY --from=builder /app/web ./web
COPY --from=builder /app/configs ./configs

# Set default config path
ENV CONFIG_PATH=/app/configs/production.json

EXPOSE 8080

ENTRYPOINT ["./analyzer"]