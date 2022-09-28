all: test

.PHONY: test
test:
	@echo "=== newrelic-integration-e2e === [ test ]: running unit tests..."
	@go test -race ./... -count=1
