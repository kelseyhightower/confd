.PHONY: build install clean test integration dep release
VERSION=`egrep -o '[0-9]+\.[0-9a-z.\-]+' version.go`
GIT_SHA=`git rev-parse --short HEAD || echo`

build:
	@echo "Building confd..."
	@mkdir -p bin
	@go build -ldflags "-X main.GitSHA=${GIT_SHA}" -o bin/confd .

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
	@for i in `find ./integration -name test.sh`; do \
		echo "Running $$i"; \
		bash $$i || exit 1; \
		bash integration/expect/check.sh || exit 1; \
		rm /tmp/confd-*; \
	done

mod:
	@go mod tidy

release:
	@docker build -q -t confd_builder -f Dockerfile .
	@for platform in darwin linux windows; do \
		if [ $$platform == windows ]; then extension=.exe; fi; \
		docker run --rm -it -v ${PWD}:/app -e "GOOS=$$platform" -e "GOARCH=amd64" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-$$platform-amd64$$extension; \
		docker run --rm -it -v ${PWD}:/tmp hairyhenderson/upx:3.96 -q /tmp/bin/confd-${VERSION}-$$platform-amd64$$extension; \
	done
	@docker run --rm -it -v ${PWD}:/app -e "GOOS=linux" -e "GOARCH=arm64" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-linux-arm64; \
	docker run --rm -it -v ${PWD}:/tmp hairyhenderson/upx:3.94 -q /tmp/bin/confd-${VERSION}-linux-arm64;
	@docker run --rm -it -v ${PWD}:/app -e "GOOS=linux" -e "GOARCH=arm" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-linux-arm32; \
	docker run --rm -it -v ${PWD}:/tmp hairyhenderson/upx:3.94 -q /tmp/bin/confd-${VERSION}-linux-arm32;
