# Testing Guide for SRELib

This guide explains the comprehensive testing strategy for the SRELib SDK, with a focus on testing OCM (OpenShift Cluster Manager) API integrations.

## Overview

SRELib uses a **three-tier testing strategy** with a unique dual-mode E2E testing approach:

1. **Unit Tests** - Fast, isolated tests with no external dependencies
2. **Integration Tests** - Mock server tests with real OCM API fixtures
3. **E2E Tests** - Dual-mode tests that work both locally (mock) and in CI (real OCM)

## Quick Start

```bash
# Run all fast tests (unit + integration, no credentials needed)
make test

# Run only unit tests
make test-unit

# Run only integration tests with mock server
make test-integration

# Run E2E tests in mock mode (no credentials needed)
make test-e2e

# Run E2E tests against real OCM (requires credentials)
export OCM_TOKEN=$(ocm token)
export TEST_CLUSTER_ID="your-test-cluster-id"
export OCM_URL=staging  # optional, defaults to staging
make test-e2e

# Run everything
make test-all
```

## Test Tier Details

### 1. Unit Tests (Fast, No Dependencies)

**Location**: `internal/*/` packages
**Build tag**: None
**Run with**: `make test-unit`

Tests internal implementation details with no external dependencies.

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

### 2. Integration Tests (Mock Server, No Credentials)

**Location**: `sdk/v1/*_test.go` (without build tags)
**Build tag**: `integration` (optional, implied by default)
**Run with**: `make test-integration`

Tests the SDK using a **mock HTTP server** that simulates OCM API responses with real fixtures.

**Key characteristics:**
- No real OCM credentials required
- Uses `setupMockServer()` to create local HTTP server
- Fast and deterministic
- Safe to run in any environment
- Tests the **entire stack** except the final HTTP call

**Architecture Flow:**
```
Test Code
   ↓
RPCClient (sdk/v1/client.go)
   ↓
RPC Server (sdk/v1/server.go)
   ↓
i1.Client (internal/i1/types.go)
   ↓
OCM functions (internal/ocm/*.go)
   ↓
OCM SDK (github.com/openshift-online/ocm-sdk-go)
   ↓
HTTP Request → Mock Server (returns real fixture data)
```

**Running:**
```bash
make test-integration
# or
go test ./sdk/v1 -v
```

**Example:**
```go
// Integration test example
func TestGetCluster_MockOCM(t *testing.T) {
    // Creates a local HTTP server with fixtures
    mockServer := setupMockServer(t)
    defer mockServer.Close()

    // Points client at mock server
    t.Setenv("OCM_URL", mockServer.URL)
    client := setupTestClient(t)

    // Makes HTTP calls to localhost, not real OCM
    cluster, err := client.GetCluster(testClusterID)
    // ...
}
```

**Note:** Full mock-based testing is necessary because the OCM SDK (`ocm-sdk-go`) manages its own HTTP client internally and doesn't expose injection points.

### 3. E2E Tests (Dual Mode: Mock or Real OCM)

**Location**: `sdk/v1/client_e2e_test.go`
**Build tag**: `e2e` (required)
**Run with**: `make test-e2e` or `go test -tags=e2e ./sdk/v1`

**Unique feature**: These tests automatically adapt based on whether credentials are available:
- **Mock mode** (no `OCM_TOKEN`): Uses mock server with fixtures for local development
- **Real mode** (`OCM_TOKEN` set): Tests against real OCM staging/production

This allows developers to run E2E tests locally without credentials, while CI runs the same tests against real OCM for validation.

#### Local Development (Mock Mode)
```bash
# No credentials needed - uses mock server
unset OCM_TOKEN
go test -tags=e2e ./sdk/v1 -v

# Test output will show:
# Running E2E test in MOCK mode (no OCM_TOKEN found)
```

**Benefits:**
- Faster feedback loop - no real API calls
- Offline development - works without VPN or internet
- No credential management - no tokens to rotate
- Deterministic tests - same fixtures, same results
- Safe experimentation - can't accidentally affect real clusters

#### CI/Production (Real OCM Mode)
```bash
# Set up credentials for real OCM
export OCM_TOKEN=$(ocm token)
export TEST_CLUSTER_ID="your-real-cluster-id"
export OCM_URL=staging  # optional, defaults to staging

# Run E2E tests against real OCM
make test-e2e
# or
go test -tags=e2e ./sdk/v1 -v

# Test output will show:
# Running E2E test in REAL OCM mode
```

**Benefits:**
- Validates against real API - catches integration issues
- Flexible configuration - can run in mock mode if secrets not configured
- Same test code - no separate test suites to maintain
- Clear mode indication - logs show which mode is active

**Safety:**
- E2E tests default to `OCM_URL=staging` to prevent accidental production queries
- Tests use read-only operations (no cluster creation/deletion)
- Some tests gracefully handle permission errors

#### How Dual Mode Works

```go
//go:build e2e
// +build e2e

type e2eTestContext struct {
    client        *i1.Client      // OCM client
    mockServer    *MockOCMServer  // Mock server (only in mock mode)
    testClusterID string          // Test cluster ID
    isRealOCM     bool           // Mode indicator
}

func TestGetCluster_E2E(t *testing.T) {
    // Auto-detects mode based on OCM_TOKEN presence
    ctx := setupE2ETest(t)
    defer ctx.Close()

    // Works in both modes:
    // - Mock: Uses fixture from testdata/fixtures/cluster_osd.json
    // - Real: Makes actual API call to OCM staging/production
    cluster, err := ctx.client.GetCluster(ctx.testClusterID)

    assert.NoError(t, err)
    assert.NotNil(t, cluster)
    // ...
}
```

**Mode detection logic:**
- If `OCM_TOKEN` is set → Real OCM mode
- If `OCM_TOKEN` is empty → Mock mode with fixtures
- No code changes needed - tests automatically adapt

## Test Coverage

The E2E test suite covers:
- `TestGetCluster_E2E` - Cluster retrieval
- `TestGetClusterAnyStatus_E2E` - Cluster retrieval regardless of status
- `TestGetClusters_E2E` - Batch cluster retrieval
- `TestIsClusterCCS_E2E` - CCS flag detection
- `TestIsHostedCluster_E2E` - Hypershift/HCP detection
- `TestGetSubscription_E2E` - Subscription lookup
- `TestGetOrgFromClusterID_E2E` - Organization lookup
- `TestGetAWSAccountIdForCluster_E2E` - AWS account ID extraction
- `TestConnectionLifecycle_E2E` - Full connection lifecycle

## Test Client Setup

All tests use `setupTestClient(t)` which creates a client via `NewTestClient()`:

```go
// sdk/v1/testing_helpers.go
func setupTestClient(t *testing.T) *i1.Client {
    logger := hclog.New(&hclog.LoggerOptions{
        Level: hclog.Error,
    })

    // Creates client using environment variables
    client, err := i1.NewTestClient(logger)
    require.NoError(t, err)
    return client
}
```

```go
// internal/i1/types.go
func NewTestClient(logger hclog.Logger) (*Client, error) {
    // Uses OCM_URL and OCM_TOKEN from environment
    conn, err := ocm.CreateTestConnection()
    if err != nil {
        return nil, err
    }

    return &Client{
        Logger:  logger,
        ocmConn: conn,
    }, nil
}
```

```go
// internal/ocm/connection_testing.go
func CreateTestConnection() (*sdk.Connection, error) {
    urlEnv := os.Getenv("OCM_URL")      // e.g., "staging", "http://localhost:8080", etc.
    tokenEnv := os.Getenv("OCM_TOKEN")   // Auth token

    // Resolves URL alias (staging → real URL) or uses URL directly
    ocmApiURL := resolveURL(urlEnv)

    return sdk.NewConnectionBuilder().
        URL(ocmApiURL).
        TokenURL(ocmApiURL).
        Tokens(tokenEnv).
        Agent("srelib-...").
        Build()
}
```

## The Key Difference Between Test Types

| Test Type | OCM_URL | OCM_TOKEN | HTTP Destination |
|-----------|---------|-----------|------------------|
| **Unit** | - | - | No HTTP calls |
| **Integration** | `http://localhost:xxxxx` (mock server) | `"fake-token"` | Local mock HTTP server |
| **E2E (Mock)** | `http://localhost:xxxxx` (mock server) | `""` (empty) | Local mock HTTP server |
| **E2E (Real)** | `"staging"` (resolves to real URL) | Real token from `ocm token` | Real OCM staging API |

**Integration tests:**
```go
mockServer := setupMockServer(t)  // Creates localhost:random-port
t.Setenv("OCM_URL", mockServer.URL)  // Points to localhost
t.Setenv("OCM_TOKEN", "fake-token")
// → HTTP calls go to localhost mock server
```

**E2E tests (mock mode):**
```go
// OCM_TOKEN not set in environment
ctx := setupE2ETest(t)  // Creates mock server automatically
// → HTTP calls go to localhost mock server
```

**E2E tests (real mode):**
```go
t.Setenv("OCM_URL", "staging")  // Resolves to https://api.stage.openshift.com
// OCM_TOKEN already set in environment with real token
// → HTTP calls go to real OCM staging
```

## Test Fixtures

**Location**: `sdk/v1/testdata/fixtures/`

Test fixtures provide real OCM API response formats captured from actual API calls and sanitized for testing.

### Available Fixtures

- `cluster_osd.json` - OpenShift Dedicated cluster
- `cluster_ccs.json` - CCS (Customer Cloud Subscription) cluster
- `cluster_hypershift.json` - Hypershift/HCP cluster
- `cluster_rosa.json` - ROSA cluster
- `subscription_response.json` - Subscription data
- `organization_response.json` - Organization data

### Loading Fixtures

```go
// In integration tests
mockServer := setupMockServer(t)
clusterID, err := mockServer.LoadClusterFromFixture("cluster_osd.json")
require.NoError(t, err)

// Or add raw JSON for custom scenarios
customJSON := json.RawMessage(`{"id": "custom-123", "name": "custom", ...}`)
mockServer.AddClusterRaw("custom-123", customJSON)
```

### Capturing New Fixtures

```bash
# Set credentials
export OCM_URL=staging
export OCM_TOKEN=$(ocm token)

# Capture cluster
ocm get /api/clusters_mgmt/v1/clusters/CLUSTER_ID > temp.json

# ⚠️ MANDATORY: Sanitize before committing (see below)

# Save to fixtures
mv sanitized.json sdk/v1/testdata/fixtures/cluster_new.json
```

### ⚠️ MANDATORY: Sanitize Fixtures

**NEVER commit fixtures with real customer data!**

Before adding any fixture:

1. Replace all real cluster IDs → `test-cluster-123`
2. Replace all AWS account IDs → `123456789012`
3. Replace all customer names → generic test names
4. Replace all real domains → `.test.t1.openshiftapps.com`
5. Remove any tokens, credentials, ARNs with real account IDs
6. Add `_metadata.sanitized: true`

See [sdk/v1/testdata/fixtures/README.md](sdk/v1/testdata/fixtures/README.md) for complete sanitization guide.

## Mock Server Features

The mock HTTP server mimics OCM API behavior:

### Supported Endpoints

- `GET /api/clusters_mgmt/v1/clusters` - List clusters (with search)
- `GET /api/clusters_mgmt/v1/clusters/{id}` - Get cluster by ID
- `GET /api/accounts_mgmt/v1/subscriptions` - List subscriptions (with search)
- `GET /api/accounts_mgmt/v1/organizations/{id}` - Get organization
- Authentication endpoints (returns dummy tokens)

### Search Syntax Support

The mock server parses OCM search syntax:

```go
// OCM search format: "id = 'value' or name = 'value'"
// Works in tests automatically
cluster, err := client.GetCluster("test-cluster-234")  // by ID
cluster, err := client.GetCluster("test-cluster-234s") // by name
```

### Automatic Indexing

Fixtures are automatically indexed by:
- Cluster: `id`, `name`, `external_id`
- Subscription: `id`, `cluster_id`, `display_name`
- Organization: `id`

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OCM_TOKEN` | Yes (E2E real mode only) | - | OCM API authentication token |
| `OCM_URL` | No | `staging` | OCM environment: `staging`, `integration`, or `production` |
| `TEST_CLUSTER_ID` | Yes (E2E real mode only) | - | ID of a known test cluster in the target environment |

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

## Running All Tests

```bash
# Fast tests only (unit + integration with mocks)
make test

# Everything including E2E
make test-all

# Or individually
make test-unit
make test-integration
make test-e2e
```

## Coverage

Check test coverage:

```bash
# Generate coverage report
go test ./sdk/v1 -cover -v

# Generate HTML coverage report
go test ./sdk/v1 -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage with E2E tests
go test -tags=e2e ./sdk/v1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Adding New Tests

### Integration Test (Mock Server)

```go
func TestMyFeature_MockOCM(t *testing.T) {
    mockServer := setupMockServer(t)
    defer mockServer.Close()

    // Register mock responses
    mockServer.RegisterResponse("/api/clusters_mgmt/v1/clusters/abc",
        loadFixture(t, "cluster.json"))

    t.Setenv("OCM_URL", mockServer.URL)
    t.Setenv("OCM_TOKEN", "fake-token")
    client := setupTestClient(t)
    defer client.Close()

    // Test your feature
    result, err := client.MyFeature("abc")
    assert.NoError(t, err)
    // ...
}
```

### E2E Test (Dual Mode)

```go
//go:build e2e
// +build e2e

func TestMyFeature_E2E(t *testing.T) {
    // Step 1: Setup (auto-detects mode)
    ctx := setupE2ETest(t)
    defer ctx.Close()

    // Step 2: (Optional) Load mock fixtures if needed for mock mode
    if !ctx.isRealOCM {
        ctx.mockServer.LoadClusterFromFixture("cluster_special.json")
    }

    // Step 3: Test your feature (same code for both modes)
    result, err := ctx.client.MyFeature(ctx.testClusterID)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Test Helpers

**Location**: `sdk/v1/testing_helpers.go`

Shared test utilities available to all test files:

- `setupTestClient(t)` - Creates a test client with environment-based OCM URL
- `loadFixture(t, filename)` - Loads JSON fixture from testdata/fixtures/
- `assertClusterValid(t, cluster)` - Validates cluster object structure
- `assertSubscriptionValid(t, sub)` - Validates subscription object structure
- `assertOrganizationValid(t, org)` - Validates organization object structure
- `getEnvOrDefault(key, default)` - Gets environment variable with fallback

## Troubleshooting

### "OCM_TOKEN not set" - E2E tests skip
**Expected for mock mode**. This is normal when running locally. Tests will use mock server.

### "connection refused" - Integration tests fail
The mock server setup failed. Check that `setupMockServer()` is called correctly.

### "unauthorized" - E2E tests fail
Your `OCM_TOKEN` is invalid or expired. Regenerate:
```bash
export OCM_TOKEN=$(ocm token)
```

### E2E tests skip even with OCM_TOKEN set
Ensure you're running with the `e2e` build tag:
```bash
go test -tags=e2e ./sdk/v1 -v
# or
make test-e2e
```

### Tests pass locally but fail in CI
- Unit tests: Should work everywhere (no credentials needed)
- Integration tests: Should work everywhere (no credentials needed)
- E2E tests (mock): Should work everywhere (no credentials needed)
- E2E tests (real): Require GitHub secrets to be configured (see E2E_SETUP.md)

### Mock test fails with "invalid OCM_URL"
Make sure your `internal/ocm/connection.go` allows full URLs:

```go
// Should accept http:// and https:// URLs for testing
if strings.HasPrefix(urlEnv, "http://") || strings.HasPrefix(urlEnv, "https://") {
    ocmApiOverride = urlEnv
}
```

### Mock test fails with "Not logged in"
Ensure you're setting a valid dummy JWT token:

```go
t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
```

### Fixture not found
Check the fixture path:
- Fixtures must be in `sdk/v1/testdata/fixtures/`
- Use the filename only, not the full path
- File must be valid JSON

### Tests are slow
- Run only unit/integration tests: `make test` (fast, no credentials)
- E2E tests are inherently slower due to real API calls
- Run E2E tests sparingly, not on every commit

## Testing Best Practices

1. **Choose the right tier:**
   - Unit tests for pure logic with no dependencies
   - Integration tests for testing client behavior with mock server
   - E2E tests for validating against real OCM (or mock for development)

2. **Keep tests fast** - Unit and integration tests should run in milliseconds

3. **Make tests deterministic** - No random data, no time dependencies

4. **Test error paths** - Don't just test the happy path

5. **Use descriptive names** - Test names should describe what they test

6. **Don't test external libraries** - Trust the OCM SDK, test your code

7. **Clean up resources** - Use `defer` to clean up connections

8. **Default to safety** - E2E tests should default to staging, never production

9. **Follow existing patterns:**
   - Use table-driven tests for multiple test cases
   - Use `testify/require` for fatal errors (setup, preconditions)
   - Use `testify/assert` for non-fatal assertions
   - Add descriptive test names and comments

10. **E2E test checklist:**
    - Add `//go:build e2e` at the top
    - Use `setupE2ETest(t)` for dual-mode support
    - Handle both mock and real modes
    - Log useful information for debugging

11. **Always sanitize fixtures** before committing

12. **Use mock tests for CI/CD** - fast and reliable

13. **Use E2E tests occasionally** - verify against real API changes

14. **Keep fixtures up to date** - when OCM API changes

15. **Document special fixtures** - if they test edge cases

## Benefits of This Testing Strategy

✅ **True E2E coverage** - Tests all your code including OCM integration layer
✅ **No credentials needed for development** - Mock tests run without OCM access
✅ **Fast** - No network calls to real APIs for local development
✅ **Realistic** - Uses real OCM API response formats
✅ **Maintainable** - Easy to add new test scenarios
✅ **Safe** - Can't accidentally affect real clusters
✅ **Flexible** - Same tests work locally and in CI
✅ **Comprehensive** - Three tiers cover all testing needs

## Resources

- [Mock Server Implementation](sdk/v1/ocm_mock_server_test.go)
- [E2E Tests Implementation](sdk/v1/client_e2e_test.go)
- [Integration Tests](sdk/v1/client_integration_test.go)
- [Fixture Documentation](sdk/v1/testdata/fixtures/README.md)
- [Connection Code](internal/ocm/connection_testing.go)
- [GitHub Actions Setup](.github/E2E_SETUP.md)
- [testify documentation](https://github.com/stretchr/testify)
- [Go testing package](https://pkg.go.dev/testing)
- [OCM SDK documentation](https://github.com/openshift-online/ocm-sdk-go)
- [Table-driven tests in Go](https://go.dev/wiki/TableDrivenTests)
