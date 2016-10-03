FROM alpine:3.4

ENV GOPATH /go

RUN mkdir -p "$GOPATH/src/" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && \
    mkdir -p /go/src/github.com/kelseyhightower/confd

RUN apk --update add go bash && \
    ln -s /go/src/github.com/kelseyhightower/confd /app

WORKDIR /app
