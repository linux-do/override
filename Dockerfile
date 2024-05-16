FROM golang:latest

WORKDIR $GOPATH/override

ADD . $GOPATH/override

RUN go build .

EXPOSE 8181

ENTRYPOINT  ["./override"]
