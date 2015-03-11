GO = godep go
INSTALL = install
RM = rm -f

GOBUILD = $(GO) build
GOTEST = $(GO) test --cover --race

# The directory to install confd in
prefix = /usr/local
bin_dir = $(prefix)/bin

# used to reference the output directory for build artifacts
build_dir = bin

all: build

build:
	$(GOBUILD) -o $(build_dir)/confd .

clean:
	$(RM) $(build_dir)/*

install:
	$(INSTALL) -c $(build_dir)/confd $(bin_dir)/confd

test:
	$(GOTEST) ./...
