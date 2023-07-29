BIN = shiba

SRCS = $(shell find . -type f -name '*.go' -print)

$(BIN): $(SRCS) go.mod go.sum
	go mod tidy
	go build -o $(BIN) $(SRCS)

.PHONY: format
format:
	goimports -w .

.PHONY: test
test: clean $(BIN)
	go mod tidy
	go vet
	go test ./...
	ls -1 tests/*.sb | xargs -L 1 ./shiba

.PHONY: clean
clean:
	rm -f $(BIN)
