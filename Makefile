BIN = shiba

SRCS = $(shell find . -type f -name '*.go' -print)
STDSRCS = $(shell find ./std -type f -name '*.sb' -print)

$(BIN): $(SRCS) $(STDSRCS) go.mod go.sum
	go mod tidy
	go build -o $(BIN) $(SRCS)

rel-build:
	go mod tidy
	go build -o $(BIN) -ldflags="-s -w -X main.version=$$VERSION" $(SRCS)

.PHONY: install
install:
	mv ./shiba /usr/local/bin/shiba

.PHONY: format
format:
	goimports -w .

.PHONY: test
test: clean $(BIN) gotest sbtest

.PHONY: gotest
gotest: clean $(BIN)
	go mod tidy
	go vet
	go test ./...

.PHONY: sbtest
sbtest: clean $(BIN)
	ls -1 tests/*.sb | xargs -L 1 ./shiba
	ls -1 tests/**/*.sb | xargs -L 1 ./shiba

.PHONY: clean
clean:
	rm -f $(BIN)
