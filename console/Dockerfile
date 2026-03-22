FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o console .

# ---- runtime ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/console .

ENV BACKEND_URL=http://backend:8080

CMD ["./console"]
