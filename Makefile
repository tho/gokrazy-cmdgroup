.PHONY: check integration lint test vuln

check: vuln lint test integration

vuln:
	govulncheck ./...

lint:
	golangci-lint run

test:
	go test -race ./...

integration:
	sh example_test.sh
