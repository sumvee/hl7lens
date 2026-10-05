BINARY := hl7lens
PKG    := github.com/sumvee/hl7lens
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/cmd.version=$(VERSION)

.PHONY: build test vet lint install snapshot clean

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

# Full local check, as CI runs it.
lint: vet test

install:
	go install -ldflags '$(LDFLAGS)' .

# Build release artifacts locally without publishing (requires goreleaser).
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -f $(BINARY)
	rm -rf dist
