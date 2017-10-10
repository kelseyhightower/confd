.PHONY: build install clean test integration dep

build:
	@echo "Building confd..."
	@mkdir -p bin
	@go build -ldflags "-X main.GitSHA=`git rev-parse --short HEAD`" -o bin/confd .

install:
	@echo "Installing confd..."
	@install -c bin/confd /usr/local/bin/confd

clean:
	@rm -f bin/*

test:
	@echo "Running tests..."
	@go test `go list ./... | grep -v vendor/`

integration:
	@echo "Running integration tests..."
	@find ./integration -name test.sh -exec bash {} \;

dep:
	@dep ensure
	@dep prune
