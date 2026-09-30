.PHONY: all build run test vet check fmt fmt-check tidy cover race install clean help

# Go toolchain downloads need checksum verification.
export GOSUMDB := sum.golang.org

GO ?= go
BINARY ?= ttypist
WORDS ?= one two three

all: check build

# Keep build unconditional. Go's own content-hash cache makes repeated builds
# cheap and avoids stale binaries when source and binary mtimes tie.
build:
	$(GO) build -buildvcs=false -trimpath -ldflags "-s -w" -o $(BINARY) .

# Example: make run WORDS='always man good same' 
run: build
	./$(BINARY) "$(WORDS)"

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: fmt-check test vet

fmt:
	gofmt -w $$(find . -name '*.go' -type f -not -path './vendor/*')

fmt-check:
	@test -z "$$(gofmt -l .)"

tidy:
	$(GO) mod tidy

cover:
	$(GO) test -coverprofile=coverage.out ./...

race:
	$(GO) test -race ./...

install:
	$(GO) install .

clean:
	rm -f $(BINARY) coverage.out

help:
	@printf '%s\n' \
		'make run WORDS="one two three"  Build and run a session' \
		'make build                       Build ./ttypist' \
		'make check                       Format check, tests, and vet' \
		'make test                        Run Go tests' \
		'make vet                         Run go vet' \
		'make fmt                         Format Go sources' \
		'make tidy                        Update go.mod and go.sum' \
		'make cover                       Run tests with coverage output' \
		'make race                        Run tests with the race detector' \
		'make install                     Install the binary with go install' \
		'make clean                       Remove local build artifacts'
