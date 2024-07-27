FROM golang:1.22.5 AS builder

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go build -o traefik-adapter .

FROM debian:buster-slim

WORKDIR /app

COPY --from=builder /app/traefik-adapter .

CMD ["./traefik-adapter"]
