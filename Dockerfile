FROM golang:alpine AS builder

WORKDIR /app

# 安装 ca-certificates 和 upx 包
RUN apk --no-cache add ca-certificates upx

ENV GO111MODULE=on GOSUMDB=off GOPROXY=https://goproxy.cn

ADD . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static"' -o override && upx --best /app/override

FROM scratch

# 将操作系统 CA 证书复制到最终镜像
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# 复制静态二进制文件
COPY --from=builder /app/override /app/override

# 复制配置文件
# COPY --from=builder /app/config.json /app/config.json

WORKDIR /app

ENTRYPOINT ["/app/override"]

EXPOSE 8181
