.PHONY: test test-unit test-integration test-e2e test-all help

# Default target
help:
	@echo "Available targets:"
	@echo "  make test            - Run unit and integration tests (fast, no credentials needed)"
	@echo "  make test-unit       - Run only unit tests in internal packages"
	@echo "  make test-integration - Run integration tests in sdk/v1"
	@echo "  make test-e2e        - Run E2E tests (mock mode locally, real mode with OCM_TOKEN)"
	@echo "  make test-all        - Run all tests (unit + integration + E2E)"
	@echo ""
	@echo "Environment variables for E2E tests:"
	@echo "  OCM_TOKEN         - OCM API token (optional: enables real OCM mode)"
	@echo "  OCM_URL           - OCM environment (default: staging, used in real mode)"
	@echo "  TEST_CLUSTER_ID   - Test cluster ID (required for real mode)"
	@echo ""
	@echo "E2E test modes:"
	@echo "  Mock mode:  Run without OCM_TOKEN (uses local fixtures, fast)"
	@echo "  Real mode:  Set OCM_TOKEN and TEST_CLUSTER_ID (tests against OCM staging)"

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

# Run E2E tests (auto-detects mock vs real mode based on OCM_TOKEN)
test-e2e:
	@echo "Running E2E tests..."
	@if [ -z "$$OCM_TOKEN" ]; then \
		echo "OCM_TOKEN not set - running in MOCK mode (using local fixtures)"; \
		echo "To run against real OCM, export OCM_TOKEN and TEST_CLUSTER_ID"; \
	else \
		echo "OCM_TOKEN detected - running in REAL OCM mode"; \
		if [ -z "$$TEST_CLUSTER_ID" ]; then \
			echo "Warning: TEST_CLUSTER_ID not set. Tests requiring real cluster will skip."; \
			echo "Example: export TEST_CLUSTER_ID=your-cluster-id"; \
		else \
			echo "Using test cluster: $$TEST_CLUSTER_ID"; \
		fi; \
		echo "Using OCM environment: $${OCM_URL:-staging}"; \
	fi
	@OCM_URL=$${OCM_URL:-staging} go test -tags=e2e ./sdk/v1 -run "_E2E$$" -v

# Run all tests (integration + E2E)
test-all: test test-e2e
