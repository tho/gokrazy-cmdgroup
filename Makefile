.PHONY: check vuln fmt fmt-diff lint test integration

check: vuln fmt-diff lint test integration

vuln:
	govulncheck ./...

fmt:
	golangci-lint fmt ./...

fmt-diff:
	golangci-lint fmt --diff ./...

lint:
	golangci-lint run

test:
	go test -race ./...

integration:
	sh example_test.sh
