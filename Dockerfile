FROM golang:1.21-alpine AS builder

WORKDIR $GOPATH/override

ADD . $GOPATH/override

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o override

FROM alpine:latest

COPY --from=builder $GOPATH/override/override /usr/local/bin/override

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/override"]

EXPOSE 8181
