.PHONY: check lint test vuln

check: vuln lint test

vuln:
	govulncheck ./...

lint:
	golangci-lint run

test:
	go test -race ./...
