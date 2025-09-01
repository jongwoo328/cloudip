build:
	goreleaser --snapshot --clean

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -cover ./...

test-bench:
	go test -bench=. ./...

.PHONY: build test test-verbose test-coverage test-bench
