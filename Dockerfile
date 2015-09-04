FROM gliderlabs/alpine:3.2
MAINTAINER Jussi Nummelin <jussi.nummelin@gmail.com>
RUN apk-install bash go bzr git mercurial subversion openssh-client ca-certificates

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go
