FROM docker.m.daocloud.io/library/golang:1.25-alpine AS builder
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main .

FROM docker.m.daocloud.io/library/alpine:latest
WORKDIR /app
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yaml ./config/config.yaml
COPY --from=builder /app/web ./web

EXPOSE 8080
EXPOSE 6060
CMD ["./main"]