.PHONY: build test lint run fmt tidy

BIN_DIR := bin
VERTEXD := $(BIN_DIR)/vertexd
VERTEXCTL := $(BIN_DIR)/vertexctl

build: $(VERTEXD) $(VERTEXCTL)

$(VERTEXD): $(shell find cmd/vertexd internal -name '*.go' 2>/dev/null)
	go build -o $(VERTEXD) ./cmd/vertexd

$(VERTEXCTL): $(shell find cmd/vertexctl internal -name '*.go' 2>/dev/null)
	go build -o $(VERTEXCTL) ./cmd/vertexctl

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

run: build
	$(VERTEXD)
