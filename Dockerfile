FROM golang:1.22.5 AS builder

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy

RUN go build -o traefik-adapter .

FROM busybox:latest

WORKDIR /app

COPY --from=builder /app/traefik-adapter .

CMD ["./traefik-adapter"]
