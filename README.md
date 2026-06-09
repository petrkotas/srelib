# The SRE library

The SRE library is a collection of common functionality required by all SRE
components.

## Testing

The project uses a three-tier testing strategy:

1. **Unit Tests** - Fast, isolated tests with no external dependencies
2. **Integration Tests** - Test Client interface behavior and error handling
3. **E2E Tests** - Validate against real OCM staging environment (requires credentials)

### Quick Start

```bash
# Run unit and integration tests (fast, no credentials needed)
make test

# Run E2E tests against OCM staging (requires credentials)
export OCM_TOKEN="your-staging-token"
export TEST_CLUSTER_ID="your-test-cluster-id"
make test-e2e
```

### Available Test Targets

```bash
make test              # Unit + integration tests
make test-unit         # Only unit tests
make test-integration  # Only integration tests
make test-e2e          # E2E tests (requires OCM credentials)
make test-all          # All tests
```

For detailed testing documentation, see [TESTING.md](TESTING.md).

