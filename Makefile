
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development

.PHONY: test
TEST_ARGS ?= -v
TEST_TARGETS ?= ./...
test: ## Test the Go modules within this package.
	@ echo ▶️ go test $(TEST_ARGS) $(TEST_TARGETS)
	go test $(TEST_ARGS) $(TEST_TARGETS)
	@ echo ✅ success!


.PHONY: lint
LINT_TARGETS ?= ./...
lint: ## Lint Go code with the installed golangci-lint
	@ echo "▶️ golangci-lint run"
	golangci-lint run $(LINT_TARGETS)
	@ echo "✅ golangci-lint run"


check-secure:
	govulncheck ./...
