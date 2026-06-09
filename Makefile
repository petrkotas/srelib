.PHONY: test test-unit test-integration test-e2e test-all help

# Default target
help:
	@echo "Available targets:"
	@echo "  make test            - Run unit and integration tests (fast, no credentials needed)"
	@echo "  make test-unit       - Run only unit tests in internal packages"
	@echo "  make test-integration - Run integration tests in sdk/v1"
	@echo "  make test-e2e        - Run E2E tests against real OCM (requires credentials)"
	@echo "  make test-all        - Run all tests (unit + integration + E2E)"
	@echo ""
	@echo "Environment variables for E2E tests:"
	@echo "  OCM_TOKEN         - OCM API token (required)"
	@echo "  OCM_URL           - OCM environment (default: staging)"
	@echo "  TEST_CLUSTER_ID   - Test cluster ID (required)"

# Run standard tests (unit + integration, no build tags)
test:
	@echo "Running unit and integration tests..."
	go test ./... -v -race

# Run only unit tests in internal packages
test-unit:
	@echo "Running unit tests..."
	go test ./internal/... -v -race

# Run integration tests in sdk/v1
test-integration:
	@echo "Running integration tests..."
	go test ./sdk/v1 -v -race

# Run E2E tests against real OCM
test-e2e:
	@echo "Running E2E tests against OCM..."
	@if [ -z "$$OCM_TOKEN" ]; then \
		echo "Error: OCM_TOKEN not set. Export OCM_TOKEN before running E2E tests."; \
		echo "Example: export OCM_TOKEN=your-token-here"; \
		exit 1; \
	fi
	@if [ -z "$$TEST_CLUSTER_ID" ]; then \
		echo "Warning: TEST_CLUSTER_ID not set. Some tests may skip."; \
		echo "Example: export TEST_CLUSTER_ID=your-cluster-id"; \
	fi
	@echo "Using OCM environment: $${OCM_URL:-staging}"
	OCM_URL=$${OCM_URL:-staging} go test -tags=e2e ./sdk/v1 -v

# Run all tests (integration + E2E)
test-all: test test-e2e
