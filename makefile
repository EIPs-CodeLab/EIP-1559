.PHONY: build test run clean

build:
	go build -o bin/simulator cmd/simulator/main.go

test:
	go test -v ./test/...

run:
	go run cmd/simulator/main.go

run-verbose:
	go run cmd/simulator/main.go -verbose

run-congestion:
	go run cmd/simulator/main.go -blocks=20 -gas=25000000

clean:
	rm -rf bin/

fmt:
	go fmt ./...

.DEFAULT_GOAL := build