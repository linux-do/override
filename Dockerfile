# 使用官方Go镜像作为构建环境
FROM golang:1.18 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# 使用scratch作为运行环境
FROM scratch
COPY --from=builder /main /main
ENTRYPOINT ["/main"]
