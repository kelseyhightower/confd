# Installation

### Binary Download

Currently confd ships binaries for OS X and Linux 64bit systems. You can download the latest release from [GitHub](https://github.com/kelseyhightower/confd/releases)

#### OS X

```
$ wget https://github.com/kelseyhightower/confd/releases/download/v0.13.0/confd-0.13.0-darwin-amd64
```

#### Linux

Download the binary
```
$ wget https://github.com/kelseyhightower/confd/releases/download/v0.13.0/confd-0.13.0-linux-amd64
```
Move the binary to an installation path, make it executable, and add to path
```
mkdir -p /opt/confd/bin
mv confd-0.13.0-linux-amd64 /opt/confd/bin/confd
chmod +x /opt/confd/bin/confd
export PATH="$PATH:/opt/confd/bin"
```

#### Building from Source

```
$ make build
$ make install
```

#### Building from Source for Alpine Linux

Since many people are using Alpine Linux as their base images for Docker there's support to build Alpine package also. Naturally by using Docker itself. :)

```
$ docker build -t confd_builder -f Dockerfile.build.alpine .
$ docker run -ti --rm -v $(pwd):/app confd_builder make build
```
The above docker commands will produce binary in the local bin directory.

#### Build for your Image using Multi-Stage build

With multi-stage builds you can keep the whole process contained in your Dockerfile using:

```
FROM golang:1.9-alpine as confd

RUN apk add --no-cache make unzip
RUN mkdir -p /go/src/github.com/kelseyhightower/confd && \
  ln -s /go/src/github.com/kelseyhightower/confd /app

WORKDIR /app

RUN wget -O /tmp/confd.zip https://github.com/kelseyhightower/confd/archive/v0.13.0.zip && \
    unzip -d /tmp/confd /tmp/confd.zip && \
    cp -r /tmp/confd/*/* /app && \
    rm -rf /tmp/confd* && \
    make build

FROM tomcat:8.5.15-jre8-alpine

COPY --from=confd /app/bin/confd /usr/local/bin/confd

# Then do other useful things...
```

### Next Steps

Get up and running with the [Quick Start Guide](quick-start-guide.md).
