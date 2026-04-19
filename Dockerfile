FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o bin/analyzer \
    ./cmd/server/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/analyzer .
COPY --from=builder /app/web ./web

# ADD THIS (quan trọng nhất)
COPY --from=builder /app/configs ./configs

EXPOSE 8080

ENTRYPOINT ["./analyzer"]