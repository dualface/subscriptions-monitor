.PHONY: build test vet clean

build:
	go build -o bin/sub-mon ./cmd/sub-mon

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

.DEFAULT_GOAL := build
