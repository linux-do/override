# 使用官方Go镜像作为构建环境
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /override

FROM alpine:latest as certs
RUN apk --update add ca-certificates

# 使用alpine作为运行环境
FROM alpine
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /override /override
ENTRYPOINT ["/override"]