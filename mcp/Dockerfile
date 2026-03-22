FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mcp ./cmd/main

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/mcp .

EXPOSE 3000

CMD ["./mcp"]
