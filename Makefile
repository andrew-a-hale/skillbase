.PHONY: build test clean lint release install coverage

BINARY_NAME=skillbase
GO=go
GORELEASER=goreleaser

build:
	$(GO) build -o $(BINARY_NAME) .

test:
	$(GO) test -v ./...

coverage:
	$(GO) test -coverprofile=coverage.txt -covermode=atomic ./...
	$(GO) tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html

install: test
	$(GO) install .

release: install
	$(GORELEASER) release --clean

release-snapshot: install
	$(GORELEASER) release --snapshot --clean
