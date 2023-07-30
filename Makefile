BIN = shiba

SRCS = $(shell find . -type f -name '*.go' -print)

$(BIN): $(SRCS) go.mod go.sum
	go mod tidy
	go build -o $(BIN) -ldflags="-s -w" $(SRCS)

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

.PHONY: rel
rel: clean $(BIN)
	mv ./shiba ./_shiba
	mkdir -p shiba/bin
	mv ./_shiba shiba/bin/shiba
	cp -r ./std shiba/
	tar -czf shiba0.0.0.linux_amd64.tar.gz shiba
	rm -rf shiba
