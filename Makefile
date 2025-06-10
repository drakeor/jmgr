.PHONY: generate vet test monitord local local-analyze local-and-analyze analyze_badger tools all tui tui-run

PROTO_DIR=api

BIN_DIR := bin
$(BIN_DIR):
	@mkdir -p $@

# Source files for monitord binary
SRC := $(shell find cmd/monitord -type f -name '*.go')

# Build tui binary
$(BIN_DIR)/tui: $(BIN_DIR) cmd/tui/main.go go.mod go.sum
	@echo "↻ building $@"
	@go build -o $@ ./cmd/tui

tui: $(BIN_DIR)/tui

# Build monitord binary
$(BIN_DIR)/monitord: $(BIN_DIR) $(SRC) go.mod go.sum
	@echo "↻ building $@"
	@go build -o $@ ./cmd/monitord

monitord: $(BIN_DIR)/monitord

# Build analyze_badger binary
$(BIN_DIR)/analyze_badger: $(BIN_DIR) cmd/analyze_badger/main.go go.mod go.sum
	@echo "↻ building $@"
	@go build -o $@ ./cmd/analyze_badger

analyze_badger: $(BIN_DIR)/analyze_badger

# Run the text-based UI
tui-run: tui
	@./bin/tui

# Run three local nodes (uses the helper script)
local: monitord
	@scripts/run-local.sh

# Analyze Badger stores after local cluster shutdown
local-analyze: analyze_badger
	@echo "⇢ Analyzing BadgerDBs in bin/node1, bin/node2, bin/node3"
	@$(BIN_DIR)/analyze_badger bin/node1 bin/node2 bin/node3

# One-liner: run cluster, then analyze after it exits
local-and-analyze: local local-analyze

local-purge:
	@bash ./scripts/purge-local.sh

# Install tool deps locally (no-op if already present)
TOOLS := \
	google.golang.org/protobuf/cmd/protoc-gen-go \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc

tools:
	@for t in $(TOOLS); do \
	  go install $$t@latest ;\
	done

# Generate Go stubs from .proto
generate: tools
	go generate ./$(PROTO_DIR)

# Run `go vet` on all packages
vet:
	go vet ./...

# Run tests with the race detector
test:
	go test -v -race ./...

# Convenience all‑in‑one target
all: generate vet test
