FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

# ---- test runner (used by docker compose --profile test) ----
FROM golang:1.24-alpine AS test

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["sh", "-c", "go test -p 1 ./... 2>&1 | grep -v '^?'"]

# ---- standalone registry exporter (no server or migrations) ----
FROM golang:1.24-alpine AS export-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o export-participants ./cmd/export-participants

FROM alpine:3.20 AS export-participants

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=export-builder /app/export-participants .
ENTRYPOINT ["/app/export-participants"]

# ---- runtime ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
