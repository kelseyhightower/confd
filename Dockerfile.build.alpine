FROM golang:1.10.2-alpine

RUN apk add --no-cache make git
RUN mkdir -p /go/src/github.com/kelseyhightower/confd && \
  ln -s /go/src/github.com/kelseyhightower/confd /app

WORKDIR /app
