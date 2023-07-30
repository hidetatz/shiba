BIN = shiba

SRCS = $(shell find . -type f -name '*.go' -print)

$(BIN): $(SRCS) go.mod go.sum
	go mod tidy
	go build -o $(BIN) -ldflags="-s -w" $(SRCS)

.PHONY: install
install:
	sudo mv ./shiba /usr/local/bin/shiba
	sudo mkdir -p /usr/lib/shiba/
	sudo cp -r ./std/* /usr/lib/shiba/

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

.PHONY: clean
clean:
	rm -f $(BIN)
