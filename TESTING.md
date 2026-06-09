# Testing Guide for srelib

This document describes the testing strategy and how to run tests for the srelib project.

## Overview

The srelib project uses a **three-tier testing strategy**:

1. **Unit Tests** - Test individual functions with no external dependencies
2. **Integration Tests** - Test Client interface with mocked/stubbed dependencies, fast and deterministic
3. **E2E Tests** - Validate against real OCM staging environment, require credentials

## Quick Start

```bash
# Run all unit and integration tests (no credentials needed)
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run E2E tests (requires OCM credentials)
export OCM_TOKEN="your-staging-token"
export OCM_URL="staging"
export TEST_CLUSTER_ID="your-test-cluster-id"
make test-e2e

# Run everything
make test-all
```

## Test Tier Details

### 1. Unit Tests

**Location:** `internal/ocm/*_test.go`

**Purpose:** Test individual utility functions in isolation with no external dependencies.

**Examples:**
- `TestGetOCMConfigLocation` - Tests OCM config file location resolution
- `TestValidateAndResolveOcmUrl` - Tests URL alias validation
- `TestGenerateQuery` - Tests query generation logic

**Running:**
```bash
make test-unit
# or
go test ./internal/... -v
```

**Characteristics:**
- No external dependencies (no OCM API calls)
- Fast execution (milliseconds)
- Always deterministic
- Run on every commit

### 2. Integration Tests

**Location:** `sdk/v1/client_integration_test.go`

**Purpose:** Test the Client interface behavior with error handling and connection lifecycle.

**Coverage:**
- Connection lifecycle (create, get, close)
- Error handling when connection not initialized
- Client method signatures and return types

**Running:**
```bash
make test-integration
# or
go test ./sdk/v1 -v
```

**Characteristics:**
- No credentials required
- Tests error paths and edge cases
- Fast execution
- Run on every commit

**Note:** Full mock-based integration testing (with httptest) is not feasible because the OCM SDK (`ocm-sdk-go`) manages its own HTTP client internally and doesn't expose injection points. For comprehensive API-level testing, see E2E tests.

### 3. E2E Tests

**Location:** `sdk/v1/client_e2e_test.go` (build tag: `e2e`)

**Purpose:** Validate that the Client interface works correctly against real OCM staging environment.

**Coverage:**
- `TestGetCluster_RealOCM` - Cluster retrieval
- `TestGetClusterAnyStatus_RealOCM` - Cluster retrieval regardless of status
- `TestGetClusters_RealOCM` - Batch cluster retrieval
- `TestIsClusterCCS_RealOCM` - CCS flag detection
- `TestIsHostedCluster_RealOCM` - Hypershift/HCP detection
- `TestGetSubscription_RealOCM` - Subscription lookup
- `TestGetOrgFromClusterID_RealOCM` - Organization lookup
- `TestGetAWSAccountIdForCluster_RealOCM` - AWS account ID extraction
- `TestConnectionLifecycle_RealOCM` - Full connection lifecycle

**Prerequisites:**
1. OCM API token for staging environment
2. Known test cluster ID in staging
3. Appropriate permissions to query OCM APIs

**Running:**
```bash
# Set required environment variables
export OCM_TOKEN="your-staging-token-here"
export TEST_CLUSTER_ID="your-test-cluster-id"

# Optional: specify OCM environment (defaults to staging)
export OCM_URL="staging"  # or "integration", "production" (be careful!)

# Run E2E tests
make test-e2e

# Or use go test directly
go test -tags=e2e ./sdk/v1 -v
```

**Characteristics:**
- Requires valid OCM credentials
- Makes real API calls to OCM
- Slower execution (seconds)
- May have transient failures (network, rate limits)
- Should default to **staging** environment for safety

**Safety:**
- E2E tests default to `OCM_URL=staging` to prevent accidental production queries
- Tests use read-only operations (no cluster creation/deletion)
- Some tests gracefully handle permission errors

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OCM_TOKEN` | Yes (E2E only) | - | OCM API authentication token |
| `OCM_URL` | No | `staging` | OCM environment: `staging`, `integration`, or `production` |
| `TEST_CLUSTER_ID` | Yes (E2E only) | - | ID of a known test cluster in the target environment |

**Getting OCM Token:**
```bash
# Using ocm CLI
ocm token

# Or from config
cat ~/.ocm.json | jq -r '.access_token'
```

**Finding Test Cluster ID:**
```bash
# List clusters in staging
ocm list clusters --url staging

# Get specific cluster details
ocm describe cluster <cluster-name> --url staging
```

## Test Fixtures

**Location:** `sdk/v1/testdata/fixtures/*.json`

Test fixtures provide sample OCM API responses for documentation and potential future use:

- `cluster_response.json` - Standard cluster response
- `cluster_ccs.json` - Customer Cloud Subscription (CCS) enabled cluster
- `cluster_hypershift.json` - Hypershift/HCP hosted cluster
- `subscription_response.json` - Subscription response
- `organization_response.json` - Organization response

**Note:** These fixtures are currently for reference. The OCM SDK's architecture makes it difficult to use these for mocking without significant refactoring.

## Test Helpers

**Location:** `sdk/v1/testing_helpers.go`

Shared test utilities available to all test files:

- `setupTestClient(t, ocmURL)` - Creates a test client with specified OCM URL
- `loadFixture(t, filename)` - Loads JSON fixture from testdata/fixtures/
- `assertClusterValid(t, cluster)` - Validates cluster object structure
- `assertSubscriptionValid(t, sub)` - Validates subscription object structure
- `assertOrganizationValid(t, org)` - Validates organization object structure
- `getEnvOrDefault(key, default)` - Gets environment variable with fallback

## Coverage

Check test coverage:

```bash
# Generate coverage report
go test ./sdk/v1 -cover -v

# Generate HTML coverage report
go test ./sdk/v1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Troubleshooting

### E2E Tests Skip with "OCM_TOKEN not set"

**Problem:** E2E tests skip even though you set `OCM_TOKEN`.

**Solution:** Ensure you're running with the `e2e` build tag:
```bash
go test -tags=e2e ./sdk/v1 -v
# or
make test-e2e
```

### E2E Tests Fail with "connection refused"

**Problem:** Cannot connect to OCM API.

**Solution:** Verify your token is valid and not expired:
```bash
# Test token
ocm whoami --url staging

# Refresh token if needed
ocm login --url staging --token <new-token>
```

### E2E Tests Fail with Permission Errors

**Problem:** "forbidden" or "not authorized" errors.

**Solution:** Some tests require specific permissions (e.g., organization access). These tests will log warnings but should not fail the overall test run. Check test output for permission-related skips.

### Tests Are Slow

**Problem:** Tests take a long time to run.

**Solution:**
- Run only unit/integration tests: `make test` (fast, no credentials)
- E2E tests are inherently slower due to real API calls
- Run E2E tests sparingly, not on every commit

## Contributing New Tests

When adding new tests:

1. **Choose the right tier:**
   - Unit tests for pure logic with no dependencies
   - Integration tests for error handling and lifecycle
   - E2E tests for validating against real OCM

2. **Follow existing patterns:**
   - Use table-driven tests for multiple test cases
   - Use `testify/require` for fatal errors (setup, preconditions)
   - Use `testify/assert` for non-fatal assertions
   - Add descriptive test names and comments

3. **E2E test checklist:**
   - Add `//go:build e2e` at the top
   - Skip if `OCM_TOKEN` not set
   - Default to `staging` environment
   - Handle permission errors gracefully
   - Log useful information for debugging

4. **Run all tests before committing:**
```bash
make test        # Fast tests
make test-e2e    # If you have credentials
```

## Testing Best Practices

1. **Keep tests fast** - Unit and integration tests should run in milliseconds
2. **Make tests deterministic** - No random data, no time dependencies
3. **Test error paths** - Don't just test the happy path
4. **Use descriptive names** - Test names should describe what they test
5. **Don't test external libraries** - Trust the OCM SDK, test your code
6. **Clean up resources** - Use `defer` to clean up connections
7. **Default to safety** - E2E tests should default to staging, never production

## Next Steps

See [TESTING_NEXT_STEPS.md](TESTING_NEXT_STEPS.md) for:
- How to capture real test fixtures from OCM API
- Future improvements and refactoring opportunities
- CI/CD setup recommendations
- Coverage improvement strategies

## Resources

- [testify documentation](https://github.com/stretchr/testify)
- [Go testing package](https://pkg.go.dev/testing)
- [OCM SDK documentation](https://github.com/openshift-online/ocm-sdk-go)
- [Table-driven tests in Go](https://go.dev/wiki/TableDrivenTests)
