FROM golang:alpine AS builder

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o override

FROM alpine:latest

COPY --from=builder /app/override /usr/local/bin/override

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/override"]

EXPOSE 8181
