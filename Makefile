BINARY := swims
GO ?= go

.PHONY: help build run test fmt vet tidy clean install

help:
	@printf "Available targets:\n"
	@printf "  build    Build the $(BINARY) binary\n"
	@printf "  run      Run the CLI with ARGS='...'\n"
	@printf "  test     Run all tests\n"
	@printf "  fmt      Format Go source files\n"
	@printf "  vet      Run go vet\n"
	@printf "  tidy     Tidy Go module files\n"
	@printf "  clean    Remove build artifacts\n"
	@printf "  install  Install the CLI with go install\n"

build:
	$(GO) build -o $(BINARY) .

run:
	$(GO) run . $(ARGS)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

clean:
	rm -f $(BINARY)

install:
	$(GO) install .
