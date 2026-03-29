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

release:
	goreleaser release --clean

clean:
	rm -rf dist/

.PHONY: build release clean test test-verbose test-coverage test-bench
