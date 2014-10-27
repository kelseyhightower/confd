PREFIX=/usr/bin

CWD=$(shell pwd)
BINNAME=$(shell basename $(CWD))

export GOPATH=$(CWD)/build
BINDIR=$(GOPATH)/bin

all: confd

$(GOPATH):
	mkdir $(GOPATH)
	mkdir -p $(GOPATH)/pkg
	mkdir -p $(BINDIR)

$(BINDIR)/godep: $(GOPATH)
	go get github.com/tools/godep

confd:  $(BINDIR)/godep
	rm -rf Godeps/_workspace/src/github.com/kelseyhightower/confd
	mkdir -p Godeps/_workspace/src/github.com/kelseyhightower/confd
	ln -s ./../../../../../../resource ./../../../../../../backends ./../../../../../../log Godeps/_workspace/src/github.com/kelseyhightower/confd/
	$(GOPATH)/bin/godep go build
	mv -f $(BINNAME) confd

install: confd
	install -D confd $(DESTDIR)/$(PREFIX)/confd

check:
	$(GOPATH)/bin/godep go test -v

clean:
	rm -rf Godeps/_workspace/src/github.com/kelseyhightower/confd
	rm -rf build confd
