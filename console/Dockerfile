FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o console .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/console .

ENTRYPOINT ["./console"]
