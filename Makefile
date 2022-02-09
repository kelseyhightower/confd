.PHONY: build install clean test integration modverify release
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

modtidy:
	@go mod tidy

modverify:
	@go mod verify

release:
	@docker build -q -t confd_builder -f Dockerfile.build.alpine .

	@for platform in darwin linux; do \
		docker run -it --rm -v ${PWD}:/app -e "GOOS=$$platform" -e "GOARCH=amd64" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-$$platform-amd64; \
		docker run -it --rm -v ${PWD}:/app -e "GOOS=$$platform" -e "GOARCH=arm64" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-$$platform-arm64; \
	done

	@docker run -it --rm -v ${PWD}:/app -e "GOOS=winddows" -e "GOARCH=amd64" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-windows-amd64.exe;
	@docker run -it --rm -v ${PWD}:/app -e "GOOS=linux" -e "GOARCH=arm" -e "CGO_ENABLED=0" confd_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/confd-${VERSION}-linux-arm32;
	@upx bin/confd-${VERSION}-*
