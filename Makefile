# Variables
VERBOSE ?= false

all: validate test

validate:
	@printf "=== newrelic-integration-e2e === [ validate ]: running golangci-lint & semgrep... "
	@cd newrelic-integration-e2e; go run -mod=readonly github.com/golangci/golangci-lint/cmd/golangci-lint run --verbose
	@[ -f .semgrep.yml ] && semgrep_config=".semgrep.yml" || semgrep_config="p/golang" ; \
	docker run --rm -v "${PWD}/newrelic-integration-e2e:/src:ro" --workdir / returntocorp/semgrep -c "$$semgrep_config"

test:
	@echo "=== newrelic-integration-e2e === [ test ]: running unit tests..."
	@cd newrelic-integration-e2e; go test -race ./... -count=1

run:
	@printf "=== newrelic-integration-e2e === [ run / $* ]: running the binary \n"
	@cd newrelic-integration-e2e; go run $(CURDIR)/newrelic-integration-e2e/cmd/main.go --root_dir=$(ROOT_DIR) \
	 --agent_dir=$(CURDIR)/agent_dir --account_id=$(ACCOUNT_ID) --api_key=$(API_KEY) --license_key=$(LICENSE_KEY) --spec_path=$(ROOT_DIR)/$(SPEC_PATH) --verbose_mode=$(VERBOSE)
